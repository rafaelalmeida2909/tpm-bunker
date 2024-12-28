from rest_framework import permissions


class IsAuthenticatedDevice(permissions.BasePermission):
    def has_permission(self, request, view):
        # Verifica se o usuário (Device) está associado à requisição
        return bool(request.user and hasattr(request.user, "uuid"))


class IsOwnerDevice(permissions.BasePermission):
    """
    Permissão que verifica se o dispositivo autenticado é o proprietário do objeto.
    """

    def has_object_permission(self, request, view, obj):
        # Verifica se o dispositivo é o proprietário do objeto
        return hasattr(request.user, "uuid") and obj.device.uuid == request.user.uuid
