from rest_framework.serializers import (
    CharField,
    DateTimeField,
    FileField,
    JSONField,
    Serializer,
    UUIDField,
)


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
