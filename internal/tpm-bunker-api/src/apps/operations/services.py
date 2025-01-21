import hashlib
from base64 import b64decode
import traceback

from bson.objectid import ObjectId
from cryptography.hazmat.primitives import hashes, serialization
from cryptography.hazmat.primitives.asymmetric import padding
from rest_framework.serializers import ValidationError
from cryptography.exceptions import InvalidSignature
from cryptography.hazmat.primitives.asymmetric import utils
from .enums import OperationTypes, StatusChoices
from .models import EncryptedPackage, Operation, OperationLog

def _verify_signature(device, encrypted_data, signature):
    try:
        # Preparação dos dados
        hashed_data = hashlib.sha256(encrypted_data).digest()
        decoded_signature = b64decode(signature)
        
        print(f"Hash para verificação (hex): {hashed_data.hex()}")
        print(f"Tamanho do hash: {len(hashed_data)}")
        print(f"Tamanho da assinatura: {len(decoded_signature)}")
        
        # Carregar a chave pública
        public_key = serialization.load_pem_public_key(
            device.public_key.encode(), 
            backend=None
        )
        
        try:
            # Verificar a assinatura usando PKCS1v15
            public_key.verify(
                decoded_signature,
                hashed_data,
                padding.PKCS1v15(),
                hashes.SHA256()
            )
            return True
        except InvalidSignature as e:
            print(f"Erro específico na verificação: {e}")
            # Tente verificar usando diferentes esquemas de padding
            try:
                public_key.verify(
                    decoded_signature,
                    hashed_data,
                    padding.PKCS1v15(),  # RSASSA-PKCS1-v1_5
                    utils.Prehashed(hashes.SHA256())  # Indica que o hash já foi calculado
                )
                print("Verificação bem sucedida com hash pré-calculado!")
                return True
            except Exception as e2:
                print(f"Tentativa alternativa também falhou: {e2}")
                return False
                
    except Exception as e:
        print(f"Erro na verificação da assinatura: {str(e)}")
        traceback.print_exc()  # Imprime o stack trace completo
        return False

class OperationService:
    def store_data(self, device, serializer_data):
        # Inicializa a operação
        operation = Operation(
            device=device,
            operation_type=OperationTypes.STORE,
            status=StatusChoices.PROCESSING,
        ).save()

        try:
            # Ler o conteúdo do arquivo criptografado
            if hasattr(serializer_data["encrypted_data"], "read"):
                encrypted_data = serializer_data["encrypted_data"].read()

            else:
                raise ValidationError({"error": "Invalid file format"})

            # Decodificar a chave simétrica
            try:
                encrypted_symmetric_key = b64decode(
                    serializer_data["encrypted_symmetric_key"]
                )
            except Exception:
                raise ValidationError(
                    {"error": "Invalid encrypted_symmetric_key format"}
                )

            # Verificar assinatura digital
            if not _verify_signature(
                device, encrypted_data, serializer_data["digital_signature"]
            ):
                raise ValidationError("Assinatura digital inválida")

            # Criar o pacote criptografado
            encrypted_package = EncryptedPackage(
                operation=operation,
                encrypted_data=encrypted_data,  # bytes
                encrypted_symmetric_key=encrypted_symmetric_key,  # bytes
                digital_signature=serializer_data["digital_signature"],
                hash_original=serializer_data["hash_original"],
                metadata=serializer_data.get("metadata", {}),
            ).save()

            # Atualizar status da operação
            operation.update(set__status=StatusChoices.COMPLETED)

            OperationLog(
                operation=operation,
                action="STORE_COMPLETED",
                details={"package_id": str(encrypted_package.id)},
            ).save()

            return {"operation_id": str(operation.id), "status": "success"}
        except Exception as e:
            operation.update(
                set__status=StatusChoices.FAILED, set__error_message=str(e)
            )
            OperationLog(
                operation=operation, action="STORE_FAILED", details={"error": str(e)}
            ).save()
            raise ValidationError({"error": str(e)})

    def retrieve_data(self, device, operation_id):
        try:
            if not ObjectId.is_valid(operation_id):
                raise ValidationError({"error": "ID de operação inválido"})

            operation = Operation.objects(
                id=operation_id, device=device, status=StatusChoices.COMPLETED
            ).first()

            if not operation:
                raise Operation.DoesNotExist()

            encrypted_package = EncryptedPackage.objects(operation=operation).first()
            if not encrypted_package:
                raise EncryptedPackage.DoesNotExist()

            return encrypted_package.encrypted_data

        except Operation.DoesNotExist:
            raise ValidationError({"error": "Operação não encontrada"}, code=404)

        except EncryptedPackage.DoesNotExist:
            raise ValidationError({"error": "Dados não encontrados"}, code=404)

        except ValidationError:
            raise

        except Exception as e:
            raise ValidationError({"error": f"Erro inesperado: {str(e)}"}, code=500)
