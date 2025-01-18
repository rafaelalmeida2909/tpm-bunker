from django.conf import settings
from rest_framework.permissions import AllowAny
from rest_framework.request import Request
from rest_framework.response import Response
from rest_framework.status import HTTP_200_OK
from rest_framework.views import APIView


class VersionView(APIView):
    serializer_class = None
    authentication_classes = []  # Removes the drf-spectacular/swagger lock
    permission_classes = [AllowAny]

    def get(self, request: Request) -> Response:
        return Response(f"TPM-BUNKER API v{settings.API_VERSION}", status=HTTP_200_OK)
