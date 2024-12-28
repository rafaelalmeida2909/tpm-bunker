from django.db.models import Q
from django_filters import CharFilter, FilterSet, OrderingFilter


class BaseFilterSet(FilterSet):
    """
    A base FilterSet class that provides a custom search filter for queryset filtering
    across multiple fields using a search term.

    The search is case-insensitive and checks if the search term is contained within
    any of the fields defined in the 'search_fields' attribute of the Meta class.

    Methods:
        filter_search_query(queryset, name, value):
            Applies a case-insensitive search across the fields specified in the
            'search_fields'
            attribute of the Meta class using the provided search value.
    """

    def filter_search_query(self, queryset, name, value):
        queries = [
            Q(**{f"{field}__icontains": value}) for field in self.Meta.search_fields
        ]

        combined_query = queries.pop()
        for query in queries:
            combined_query |= query

        queryset = queryset.filter(combined_query)

        return queryset


class OrderingFilter(OrderingFilter):
    """
    A custom ordering filter that allows the user to specify which fields to use when
    ordering the results.

    Args:
        fields (list): A list of fields that can be used for ordering.
        label (str): A custom label for the ordering filter.
    """

    def __init__(self, *args, **kwargs):
        fields = kwargs.pop("fields", [])
        label = kwargs.pop("label", "Which fields to use when ordering the results.")
        super().__init__(fields=fields, label=label, *args, **kwargs)


class SearchFilter(CharFilter):
    """
    A custom search filter that allows the user to search across multiple fields.

    This filter leverages a custom search method ('filter_search_query') to perform a
    case-insensitive search across the fields specified in 'fields'.

    Args:
        method (str): The method to use for filtering (default is 'filter_search_query').
        label (str): The label for the search filter (default is 'A search term.').
        fields (list): A list of fields that the search filter will operate on.
    """

    def __init__(self, *args, **kwargs):
        method = kwargs.pop("method", "filter_search_query")
        label = kwargs.pop("label", "A search term.")
        self.fields = kwargs.pop("fields", [])
        super().__init__(method=method, label=label, *args, **kwargs)
