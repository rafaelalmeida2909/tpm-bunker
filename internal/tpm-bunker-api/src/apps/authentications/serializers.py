from rest_framework.serializers import (
    BooleanField,
    CharField,
    DateTimeField,
    Serializer,
    UUIDField,
)


class DeviceTokenSerializer(Serializer):
    token = CharField(read_only=True)
    created_at = DateTimeField(read_only=True)
    expires_at = DateTimeField(read_only=True)
    is_revoked = BooleanField()


class LoginSerializer(Serializer):
    uuid = UUIDField(help_text="UUID único do dispositivo registrado")
    ek_certificate = CharField(
        help_text="Certificado EK (TPM Endorsement Key) do dispositivo"
    )


class LoginSerializer(Serializer):
    uuid = UUIDField(help_text="UUID único do dispositivo registrado")
    ek_certificate = CharField(
        help_text="Certificado EK (TPM Endorsement Key) do dispositivo"
    )


class TokenSerializer(Serializer):
    token = CharField(help_text="Token usado para validação ou revogação")


class TokenResponseSerializer(Serializer):
    token = CharField(help_text="Token de acesso gerado")
    expires_at = DateTimeField(help_text="Data de expiração do token")


class TokenValiditySerializer(Serializer):
    valid = BooleanField(help_text="Indica se o token é válido")


class TokenStatusSerializer(Serializer):
    status = CharField(help_text="Status da operação de revogação")
