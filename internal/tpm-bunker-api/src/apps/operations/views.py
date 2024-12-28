from django.http import HttpResponse
from drf_spectacular.utils import (
    OpenApiParameter,
    OpenApiTypes,
    extend_schema,
    extend_schema_view,
)
from rest_framework import status, viewsets
from rest_framework.decorators import action
from rest_framework.parsers import MultiPartParser
from rest_framework.permissions import IsAuthenticated
from rest_framework.response import Response

from .models import Operation
from .serializers import (
    OperationSerializer,
    RetrieveDataSerializer,
    StoreDataSerializer,
)
from .services import OperationService


@extend_schema_view(
    list=extend_schema(
        summary="Retorna uma lista de operações",
        description="Retorna uma lista de todas as operações do dispositivo autenticado.",
        parameters=[
            OpenApiParameter(
                name="X-Device-UUID",
                description="UUID único do dispositivo",
                required=True,
                type=OpenApiTypes.STR,
                location=OpenApiParameter.HEADER,
            ),
        ],
    ),
    store_data=extend_schema(
        summary="Armazena dados criptografados",
        description="""
       Armazena dados criptografados enviados pelo dispositivo.
       Valida a assinatura digital e cria um novo pacote criptografado.
       """,
        parameters=[
            OpenApiParameter(
                name="X-Device-UUID",
                description="UUID único do dispositivo",
                required=True,
                type=OpenApiTypes.STR,
                location=OpenApiParameter.HEADER,
            ),
        ],
    ),
    retrieve_data=extend_schema(
        summary="Recupera dados criptografados",
        description="""
       Recupera dados criptografados armazenados.
       Retorna o pacote criptografado associado à operação solicitada.
       """,
        parameters=[
            OpenApiParameter(
                name="X-Device-UUID",
                description="UUID único do dispositivo",
                required=True,
                type=OpenApiTypes.STR,
                location=OpenApiParameter.HEADER,
            ),
            OpenApiParameter(
                name="OperationID",
                description="ID da operação",
                required=True,
                type=OpenApiTypes.STR,
                location=OpenApiParameter.QUERY,
            ),
        ],
    ),
)
class OperationViewSet(viewsets.GenericViewSet):
    queryset = Operation.objects.all()
    serializer_class = OperationSerializer
    permission_classes = [IsAuthenticated]
    service_class = OperationService()
    parser_classes = [MultiPartParser]

    def get_serializer_class(self):
        if self.action == "store_data":
            return StoreDataSerializer
        elif self.action == "retrieve_data":
            return RetrieveDataSerializer
        return OperationSerializer

    def get_queryset(self):
        if getattr(self, "swagger_fake_view", False):
            return Operation.objects.none()
        return Operation.objects.filter(device=self.request.device)

    def list(self, request, *args, **kwargs):
        queryset = self.get_queryset()
        serializer = self.get_serializer(queryset, many=True)
        return Response(serializer.data)

    @action(detail=False, methods=["post"])
    def store_data(self, request):
        serializer = self.get_serializer(data=request.data)
        serializer.is_valid(raise_exception=True)

        response = self.service_class.store_data(
            device=request.device, serializer_data=serializer.validated_data
        )

        return Response(response, status=status.HTTP_201_CREATED)

    @action(detail=False, methods=["get"])
    def retrieve_data(self, request):
        operation_id = request.query_params.get("OperationID")
        if not operation_id:
            return Response(
                {"error": "OperationID is required"}, status=status.HTTP_400_BAD_REQUEST
            )

        file = self.service_class.retrieve_data(
            device=request.device,
            operation_id=operation_id,
        )

        response = HttpResponse(file, content_type="application/octet-stream")
        response["Content-Disposition"] = f"attachment; filename=encrypted_data.bin"
        response["Content-Transfer-Encoding"] = "binary"
        response["Accept-Ranges"] = "bytes"

        return response
