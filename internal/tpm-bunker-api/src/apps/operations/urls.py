from django.urls import include, path
from rest_framework.routers import DefaultRouter

from .views import OperationViewSet

router = DefaultRouter()

router.register("", OperationViewSet, basename="operation")

urlpatterns = [
    path("", include(router.urls)),
]
