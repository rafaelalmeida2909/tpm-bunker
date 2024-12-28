from datetime import datetime

from drf_spectacular.extensions import OpenApiAuthenticationExtension
from rest_framework.authentication import BaseAuthentication
from rest_framework.exceptions import AuthenticationFailed

from .models import DeviceToken


class DeviceProxy:
    def __init__(self, device):
        self.device = device

    @property
    def is_authenticated(self):
        # Sempre retorna True porque o dispositivo já foi autenticado
        return True

    def __getattr__(self, item):
        # Delegar atributos ao objeto Device encapsulado
        return getattr(self.device, item)


class DeviceTokenAuthentication(BaseAuthentication):
    def authenticate(self, request):
        token = request.META.get("HTTP_AUTHORIZATION")
        if not token:
            return None

        try:
            # Remove o prefixo "Bearer"
            token = token.split(" ")[1]

            # Busca o token no MongoDB
            device_token = DeviceToken.objects(
                token=token, is_revoked=False, expires_at__gt=datetime.now()
            ).first()

            if not device_token:
                raise AuthenticationFailed("Token inválido ou expirado")

            # Retorna o proxy em vez do objeto Device diretamente
            return (DeviceProxy(device_token.device), None)
        except (IndexError, AuthenticationFailed):
            raise AuthenticationFailed("Token inválido ou expirado")


class DeviceTokenAuthenticationScheme(OpenApiAuthenticationExtension):
    target_class = "authentications.backends.DeviceTokenAuthentication"
    name = "Bearer"

    def get_security_definition(self, auto_schema):
        return {
            "type": "http",
            "scheme": "bearer",
            "bearerFormat": "JWT",
            "description": "Token de autenticação do dispositivo",
        }
