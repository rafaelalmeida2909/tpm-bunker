import sys
from os import environ, path
from pathlib import Path

from dotenv import load_dotenv
from mongoengine import connect
from utils.no_migrations import DisableMigrations

load_dotenv()

API_MAJOR = "v1"
API_VERSION = "1.0"

# Build paths inside the project like this: BASE_DIR / 'subdir'.
BASE_DIR = Path(__file__).resolve().parent.parent
sys.path.insert(0, path.join(BASE_DIR, "apps"))


# Quick-start development settings - unsuitable for production
# See https://docs.djangoproject.com/en/4.2/howto/deployment/checklist/

# SECURITY WARNING: keep the secret key used in production secret!
SECRET_KEY = environ.get("SECRET_KEY")

# SECURITY WARNING: don't run with debug turned on in production!
DEBUG = True

ALLOWED_HOSTS = environ.get("ALLOWED_HOSTS", "").split(" ")

# Application definition

DJANGO_APPS = [
    "django.contrib.auth",
    "django.contrib.contenttypes",
    "django.contrib.staticfiles",
]

THIRD_PARTY_APPS = [
    "rest_framework",
    "django_filters",
    "drf_spectacular",
    "drf_spectacular_sidecar",
    "corsheaders",
]

LOCAL_APPS = ["authentications", "devices", "operations"]

INSTALLED_APPS = DJANGO_APPS + THIRD_PARTY_APPS + LOCAL_APPS

MIDDLEWARE = [
    "django.middleware.security.SecurityMiddleware",
    "corsheaders.middleware.CorsMiddleware",
    "django.middleware.common.CommonMiddleware",
    "django.middleware.csrf.CsrfViewMiddleware",
    "django.middleware.clickjacking.XFrameOptionsMiddleware",
    "authentications.middlewares.DeviceAuthenticationMiddleware",
]

ROOT_URLCONF = "core.urls"

TEMPLATES = [
    {
        "BACKEND": "django.template.backends.django.DjangoTemplates",
        "DIRS": [],
        "APP_DIRS": True,
        "OPTIONS": {
            "context_processors": [
                "django.template.context_processors.debug",
                "django.template.context_processors.request",
            ],
        },
    },
]

WSGI_APPLICATION = "core.wsgi.application"


# Database
# https://docs.djangoproject.com/en/4.2/ref/settings/#databases
MIGRATION_MODULES = DisableMigrations()

MONGO_DB_NAME = environ.get("MONGO_DB_NAME", "test")
MONGO_HOST = environ.get("MONGO_HOST", "localhost")
MONGO_PORT = int(environ.get("MONGO_PORT", 27017))
MONGO_USERNAME = environ.get("MONGO_USERNAME", "admin")
MONGO_PASSWORD = environ.get("MONGO_PASSWORD", "admin")
MONGO_AUTH_SOURCE = environ.get("MONGO_AUTH_SOURCE", "admin")

connect(
    db=MONGO_DB_NAME,
    host=MONGO_HOST,
    port=MONGO_PORT,
    username=MONGO_USERNAME,
    password=MONGO_PASSWORD,
    authentication_source=MONGO_AUTH_SOURCE,  # ou outro db, se necessário
)

# Internationalization
# https://docs.djangoproject.com/en/4.2/topics/i18n/

LANGUAGE_CODE = "en-us"

TIME_ZONE = environ.get("TIME_ZONE", "UTC")

USE_I18N = True

USE_TZ = bool(environ.get("TIME_ZONE"))


# Static files (CSS, JavaScript, Images)
# https://docs.djangoproject.com/en/4.2/howto/static-files/

STATIC_URL = "static/"

# Default primary key field type
# https://docs.djangoproject.com/en/4.2/ref/settings/#default-auto-field

DEFAULT_AUTO_FIELD = "django.db.models.BigAutoField"

REST_FRAMEWORK = {
    "DEFAULT_PAGINATION_CLASS": "utils.pagination.StandardResultsSetPagination",
    "PAGE_SIZE": 10,
    "DEFAULT_RENDERER_CLASSES": [
        "rest_framework.renderers.JSONRenderer",
    ],
    "DEFAULT_PARSER_CLASSES": [
        "rest_framework.parsers.JSONParser",
    ],
    "DEFAULT_FILTER_BACKENDS": ["django_filters.rest_framework.DjangoFilterBackend"],
    "DEFAULT_SCHEMA_CLASS": "drf_spectacular.openapi.AutoSchema",
    "EXCEPTION_HANDLER": "utils.custom_exception_handler.custom_exception_handler",
    "DEFAULT_AUTHENTICATION_CLASSES": [
        "authentications.backends.DeviceTokenAuthentication",
    ],
    "DEFAULT_PERMISSION_CLASSES": [
        "authentications.permissions.IsAuthenticatedDevice",
    ],
}

SPECTACULAR_SETTINGS = {
    "TITLE": "TPM BUNKER API",
    "DESCRIPTION": "API para gerenciamento seguro de dispositivos TPM e armazenamento de arquivos",
    "VERSION": API_VERSION,
    "SERVE_INCLUDE_SCHEMA": False,
    "SWAGGER_UI_DIST": "SIDECAR",
    "SWAGGER_UI_FAVICON_HREF": "SIDECAR",
    "REDOC_DIST": "SIDECAR",
    "SCHEMA_PATH_PREFIX": r"/api/v[0-9]",
    "SWAGGER_UI_SETTINGS": {
        "persistAuthorization": True,
        "docExpansion": "none",
    },
    "SECURITY": [{"Bearer": []}],
    "COMPONENTS": {
        "securitySchemes": {
            "Bearer": {
                "type": "http",
                "scheme": "bearer",
                "bearerFormat": "JWT",
                "description": "Token de autenticação do dispositivo. Use o header: Authorization: Bearer <token>",
            }
        }
    },
    "COMPONENT_SPLIT_REQUEST": True,
}

CORS_ALLOW_ALL_ORIGINS = True
