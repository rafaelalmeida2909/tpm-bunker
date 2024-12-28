from base64 import b64decode

from bson.objectid import ObjectId
from cryptography.hazmat.primitives import hashes, serialization
from cryptography.hazmat.primitives.asymmetric import padding
from rest_framework.serializers import ValidationError

from .enums import OperationTypes, StatusChoices
from .models import EncryptedPackage, Operation, OperationLog


def _verify_signature(device, encrypted_data, signature):
    try:
        # Get the public key
        public_key = serialization.load_pem_public_key(
            device.public_key.encode(), backend=None
        )

        # If encrypted_data is an InMemoryUploadedFile, read it
        if hasattr(encrypted_data, "read"):
            # Save current position
            pos = encrypted_data.tell()
            # Read the data
            data = encrypted_data.read()
            # Reset position for later use
            encrypted_data.seek(pos)
        else:
            data = encrypted_data

        # Verify signature diretamente sobre os dados, sem gerar hash
        public_key.verify(
            b64decode(signature),
            data,  # Aqui mudou - usa os dados diretos
            padding.PSS(
                mgf=padding.MGF1(hashes.SHA256()), salt_length=padding.PSS.MAX_LENGTH
            ),
            hashes.SHA256(),
        )
        return True
    except Exception as e:
        print(f"Erro na verificação da assinatura: {e}")  # Adicionar log para debug
        return False


class OperationService:
    def store_data(self, device, serializer_data):
        # Criar nova operação
        operation = Operation(
            device=device,
            operation_type=OperationTypes.STORE,
            status=StatusChoices.PROCESSING,
        ).save()

        try:

            try:
                encrypted_data = serializer_data["encrypted_data"].read()
            except AttributeError:
                raise ValidationError({"error": "Invalid file format"})

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
            # Criar pacote criptografado

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
