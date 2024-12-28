import secrets
from datetime import timedelta

from devices.models import Device
from django.utils import timezone
from rest_framework.serializers import ValidationError

from .models import DeviceToken
from .serializers import TokenResponseSerializer, TokenValiditySerializer


class AuthService:
    def login(self, serializer_data):
        try:
            # Validar se o dispositivo existe e está ativo
            device = Device.objects.get(
                uuid=serializer_data["uuid"],
                ek_certificate=serializer_data["ek_certificate"],
                is_active=True,
            )

            # Revogar tokens anteriores (opcional)
            DeviceToken.objects.filter(device=device, is_revoked__in=[False]).update(
                is_revoked=True
            )

            # Criar um novo token para o dispositivo
            token = self.create_token(device)

            return TokenResponseSerializer(
                {"token": token.token, "expires_at": token.expires_at}
            ).data
        except Device.DoesNotExist:
            raise ValidationError(
                {"error": "Dispositivo não encontrado ou credenciais inválidas"},
                code=400,
            )

    def verify_token(self, serializer_data):
        device_token = DeviceToken.objects.filter(
            token=serializer_data["token"],
            is_revoked=False,
            expires_at__gt=timezone.now(),
        ).first()
        return TokenValiditySerializer({"valid": bool(device_token)}).data

    def revoke_token(self, serializer_data):
        try:
            device_token = DeviceToken.objects.get(token=serializer_data["token"])
            if device_token.is_revoked:
                raise ValidationError({"token": "Token já está revogado"})

            device_token.is_revoked = True
            device_token.save()

            return TokenResponseSerializer(device_token).data
        except DeviceToken.DoesNotExist:
            raise ValidationError({"token": "Token não encontrado"}, code=404)

    def create_token(self, device):
        return DeviceToken.objects.create(
            device=device,
            token=secrets.token_urlsafe(32),
            expires_at=timezone.now() + timedelta(days=30),
        )
