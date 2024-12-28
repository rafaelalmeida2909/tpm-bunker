from rest_framework.exceptions import ErrorDetail
from rest_framework.views import exception_handler


def error_detail_to_dict(error_detail: ErrorDetail):
    """
    Converts an instance of ErrorDetail into a dictionary containing
    both the message and the error code.

    Args:
        error_detail (ErrorDetail): The ErrorDetail instance.

    Returns:
        dict: A dictionary with the error message and code.
    """
    return {"message": error_detail, "code": error_detail.code}


def treat_errors(data):
    """
    Recursively processes a nested structure (usually a dictionary of errors)
    to transform all ErrorDetail objects into dictionaries using
    the error_detail_to_dict function. It modifies the input data in place.

    Args:
        data (dict): A dictionary representing the error data, which may contain
                     ErrorDetail objects or lists of ErrorDetail objects.

    Returns:
        None: The function modifies the input dictionary in place.
    """
    for key, value in data.items():
        if isinstance(value, list):
            transformed_details = []
            for error_detail in value:
                if isinstance(error_detail, ErrorDetail):
                    transformed_detail = error_detail_to_dict(error_detail=error_detail)
                    transformed_details.append(transformed_detail)
            data[key] = transformed_details

        elif isinstance(value, ErrorDetail):
            data[key] = error_detail_to_dict(error_detail=value)

        elif isinstance(value, dict):
            treat_errors(data[key])


def custom_exception_handler(exc, context):
    """
    Custom exception handler that wraps around the default DRF (Django Rest Framework)
    exception handler to add additional processing for error details.

    - It removes the "code" key from the response data, if present.
    - It transforms any ErrorDetail objects in the response data into dictionaries.

    Args:
        exc (Exception): The exception that was raised.
        context (dict): The context in which the exception was raised.

    Returns:
        Response: The modified or original DRF response object.
    """
    response = exception_handler(exc, context)

    if response is not None:
        if "code" in response.data:
            del response.data["code"]

        treat_errors(data=response.data)

    return response
    return response
