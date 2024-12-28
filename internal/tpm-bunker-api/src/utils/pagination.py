from rest_framework.pagination import PageNumberPagination


class StandardResultsSetPagination(PageNumberPagination):
    """
    A pagination class that sets default pagination behavior for API responses.

    Attributes:
        page_size (int): The default number of items per page (default is 10).
        page_size_query_param (str): The query parameter that allows the client to set
                                     the page size.
        max_page_size (int): The maximum number of items that can be returned per page
                             (default is 100).
    """

    page_size = 10
    page_size_query_param = "page_size"
    max_page_size = 100
    max_page_size = 100
