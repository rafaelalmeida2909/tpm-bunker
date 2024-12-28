from drf_spectacular.utils import extend_schema, extend_schema_view
from rest_framework import status, viewsets
from rest_framework.decorators import action
from rest_framework.permissions import AllowAny
from rest_framework.response import Response

from .serializers import LoginSerializer, TokenSerializer
from .services import AuthService


@extend_schema_view(
    login=extend_schema(
        summary="Autenticar dispositivo",
        description="""
        Endpoint para autenticar um dispositivo e gerar um token de acesso.
        O token gerado é válido por 30 dias e pode ser usado para acessar 
        outros endpoints protegidos da API.
        """,
    ),
    verify_token=extend_schema(
        summary="Verificar validade do token",
        description="""
        Verifica se um token de acesso é válido e não está expirado.
        Retorna um boolean indicando a validade do token.
        """,
    ),
    revoke_token=extend_schema(
        summary="Revogar acesso do token",
        description="""
        Revoga um token de acesso existente.
        Uma vez revogado, o token não pode mais ser usado para autenticação.
        """,
    ),
)
class AuthViewSet(viewsets.ViewSet):
    permission_classes = [AllowAny]
    service_class = AuthService()

    def get_serializer_class(self):
        if self.action == "verify_token" or self.action == "revoke_token":
            return TokenSerializer
        return LoginSerializer

    def get_serializer(self, *args, **kwargs):
        serializer_class = self.get_serializer_class()
        return serializer_class(*args, **kwargs)

    @action(detail=False, methods=["post"])
    def login(self, request):
        serializer = self.get_serializer(data=request.data)
        serializer.is_valid(raise_exception=True)

        response = self.service_class.login(serializer_data=serializer.validated_data)

        return Response(response, status=status.HTTP_200_OK)

    @action(detail=False, methods=["post"])
    def verify_token(self, request):
        serializer = self.get_serializer(data=request.data)
        serializer.is_valid(raise_exception=True)

        response = self.service_class.verify_token(
            serializer_data=serializer.validated_data
        )

        return Response(response, status=status.HTTP_200_OK)

    @action(detail=False, methods=["post"])
    def revoke_token(self, request):
        serializer = self.get_serializer(data=request.data)
        serializer.is_valid(raise_exception=True)

        response = self.service_class.revoke_token(
            serializer_data=serializer.validated_data
        )

        return Response(response, status=status.HTTP_200_OK)
