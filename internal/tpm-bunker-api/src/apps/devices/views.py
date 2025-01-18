from drf_spectacular.utils import extend_schema, extend_schema_view
from rest_framework import status, viewsets
from rest_framework.permissions import AllowAny, IsAuthenticated
from rest_framework.response import Response
from rest_framework.serializers import ValidationError

from .models import Device
from .serializers import DeviceRegistrationSerializer, DeviceSerializer
from .services import DevicesService


@extend_schema_view(
    list=extend_schema(
        summary="List all devices",
        description="Retorna uma lista de todos os dispositivos registrados no sistema.",
    ),
    retrieve=extend_schema(
        summary="Retrieve device details",
        description="Retorna os detalhes de um dispositivo específico pelo UUID.",
    ),
    create=extend_schema(
        summary="Register device with TPM",
        description="""
       Registra um novo dispositivo com suas credenciais TPM.
       Valida o certificado EK e cria um novo registro de dispositivo.
       """,
    ),
    update=extend_schema(
        summary="Update device",
        description="Atualiza os dados de um dispositivo existente.",
    ),
    partial_update=extend_schema(
        summary="Partially update device",
        description="Atualiza parcialmente os dados de um dispositivo existente.",
    ),
    destroy=extend_schema(
        summary="Delete device", description="Remove um dispositivo do sistema."
    ),
)
class DeviceViewSet(viewsets.ModelViewSet):
    permission_classes = [IsAuthenticated]
    serializer_class = DeviceSerializer
    service_class = DevicesService()
    lookup_field = "uuid"

    def get_queryset(self):
        if getattr(self, "swagger_fake_view", False):
            return Device.objects().none()
        return Device.objects

    def get_serializer_class(self):
        if self.action == "create":
            return DeviceRegistrationSerializer
        return super().get_serializer_class()

    def get_permissions(self):
        if self.action == "create" or self.action == "retrieve":
            self.permission_classes = [AllowAny]
        else:
            self.permission_classes = [IsAuthenticated]
        return super().get_permissions()

    def create(self, request, *args, **kwargs):
        serializer = self.get_serializer(data=request.data)
        serializer.is_valid(raise_exception=True)

        try:
            device = self.service_class.register_device(
                serializer_data=serializer.validated_data
            )
            return Response(device, status=status.HTTP_201_CREATED)
        except ValidationError as e:
            return Response(e.detail, status=status.HTTP_400_BAD_REQUEST)
