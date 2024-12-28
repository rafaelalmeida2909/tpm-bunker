from django.db.models import TextChoices
from django.utils.translation import gettext_lazy as _


class OperationTypes(TextChoices):
    STORE = "STORE", "Armazenamento"
    RETRIEVE = "RETRIEVE", "Recuperação"
    DELETE = "DELETE", "Deleção"


class StatusChoices(TextChoices):
    PENDING = "PENDING", "Pendente"
    PROCESSING = "PROCESSING", "Processando"
    COMPLETED = "COMPLETED", "Completado"
    FAILED = "FAILED", "Falhou"
