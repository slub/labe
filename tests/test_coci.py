"""
Unit tests for labe.
"""

import requests
import pytest
from labe.coci import get_redirect_url


def test_get_redirct_url():
    with pytest.raises(requests.exceptions.MissingSchema):
        get_redirect_url("xxx")

    assert get_redirect_url("https://google.com") == "https://www.google.com/"
    assert get_redirect_url("http://google.com") == "https://www.google.com/?gws_rd=ssl"
    assert (
        get_redirect_url("https://doi.org/10.1111/icad.12417")
        == "https://onlinelibrary.wiley.com/doi/10.1111/icad.12417"
    )
