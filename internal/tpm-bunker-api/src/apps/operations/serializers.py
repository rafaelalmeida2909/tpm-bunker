from rest_framework.serializers import (
    CharField,
    DateTimeField,
    FileField,
    FloatField,
    JSONField,
    Serializer,
    UUIDField,
)

from .models import EncryptedPackage


class OperationSerializer(Serializer):
    id = UUIDField(read_only=True)
    device = CharField(help_text="ID do dispositivo relacionado à operação")
    operation_type = CharField(help_text="Tipo da operação")
    status = CharField(read_only=True, help_text="Status da operação")
    error_message = CharField(
        allow_null=True, required=False, help_text="Mensagem de erro (se houver)"
    )
    created_at = DateTimeField(read_only=True, help_text="Data de criação da operação")
    updated_at = DateTimeField(
        read_only=True, help_text="Última atualização da operação"
    )

    file_name = CharField(
        read_only=True, help_text="Nome do arquivo criptografado", required=False
    )
    file_size = FloatField(
        read_only=True, help_text="Tamanho do arquivo criptografado", required=False
    )

    def to_representation(self, instance):
        # Representação padrão do serializer
        representation = super().to_representation(instance)

        # Tenta buscar o pacote criptografado relacionado
        encrypted_package = EncryptedPackage.objects(operation=instance).first()
        if encrypted_package:
            representation["file_name"] = encrypted_package.file_name
            representation["file_size"] = encrypted_package.file_size
        else:
            representation["file_name"] = None
            representation["file_size"] = None

        return representation


class EncryptedPackageSerializer(Serializer):
    encrypted_data = FileField(help_text="Dados criptografados")
    encrypted_symmetric_key = CharField(help_text="Chave simétrica criptografada")
    digital_signature = CharField(help_text="Assinatura digital")
    hash_original = CharField(help_text="Hash dos dados originais")
    metadata = JSONField(required=False, help_text="Metadados adicionais")


class OperationLogSerializer(Serializer):
    id = UUIDField(read_only=True)
    operation = CharField(help_text="ID da operação relacionada")
    action = CharField(help_text="Ação registrada no log")
    details = JSONField(help_text="Detalhes da ação registrada")
    timestamp = DateTimeField(read_only=True, help_text="Data e hora do log")


class StoreDataSerializer(Serializer):
    encrypted_data = FileField(help_text="Arquivo contendo os dados criptografados")
    encrypted_symmetric_key = CharField(help_text="Chave simétrica criptografada")
    digital_signature = CharField(help_text="Assinatura digital do dispositivo")
    hash_original = CharField(help_text="Hash dos dados originais")
    metadata = JSONField(required=False, help_text="Metadados adicionais")


class RetrieveDataSerializer(Serializer):
    operation_id = CharField(help_text="ID da operação para recuperação de dados")
