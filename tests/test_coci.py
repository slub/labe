"""
Unit tests for labe. Most not mocked yet, hence slow.
"""

import collections
import socket

import pytest
import requests

from labe.coci import get_figshare_download_link, get_redirect_url


def internet(host="8.8.8.8", port=53, timeout=3):
    """
    Host: 8.8.8.8 (google-public-dns-a.google.com)
    OpenPort: 53/tcp
    Service: domain (DNS/TCP)
    """
    try:
        socket.setdefaulttimeout(timeout)
        socket.socket(socket.AF_INET, socket.SOCK_STREAM).connect((host, port))
        return True
    except socket.error as ex:
        return False


@pytest.mark.skipif(not internet(), reason="no internet")
def test_get_redirct_url():
    if not internet():
        pytest.skip("no internet")
    with pytest.raises(requests.exceptions.MissingSchema):
        get_redirect_url("undefined")

    assert get_redirect_url("https://google.com") == "https://www.google.com/"
    assert get_redirect_url("http://google.com") == "https://www.google.com/?gws_rd=ssl"
    assert (
        get_redirect_url("https://doi.org/10.1111/icad.12417")
        == "https://onlinelibrary.wiley.com/doi/10.1111/icad.12417"
    )


@pytest.mark.skipif(not internet(), reason="no internet")
def test_get_figshare_download_link():
    Case = collections.namedtuple("Case", "link result")
    cases = (
        Case(
            "https://doi.org/10.6084/m9.figshare.6741422.v11",
            "https://figshare.com/ndownloader/articles/6741422/versions/11",
        ),
        Case(
            "https://doi.org/10.6084/m9.figshare.6741422.v7",
            "https://figshare.com/ndownloader/articles/6741422/versions/7",
        ),
    )
    for c in cases:
        assert get_figshare_download_link(c.link) == c.result
