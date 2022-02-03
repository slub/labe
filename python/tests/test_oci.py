"""
Unit tests for labe. Most not mocked yet, hence slow.
"""

import collections
import socket

import pytest
import requests

from labe.oci import get_figshare_download_link, get_terminal_url


def no_internet(host="8.8.8.8", port=53, timeout=3):
    """
    Host: 8.8.8.8 (google-public-dns-a.google.com)
    OpenPort: 53/tcp
    Service: domain (DNS/TCP)
    """
    try:
        socket.setdefaulttimeout(timeout)
        socket.socket(socket.AF_INET, socket.SOCK_STREAM).connect((host, port))
        return False
    except socket.error as ex:
        return True


@pytest.mark.skipif(no_internet(), reason="no internet")
def test_get_redirct_url():
    with pytest.raises(requests.exceptions.MissingSchema):
        get_terminal_url("undefined")

    assert get_terminal_url("https://google.com") == "https://www.google.com/"
    assert get_terminal_url("http://google.com") == "https://www.google.com/?gws_rd=ssl"
    assert (get_terminal_url("https://doi.org/10.1111/icad.12417") == "https://onlinelibrary.wiley.com/doi/10.1111/icad.12417")


@pytest.mark.skipif(no_internet(), reason="no internet")
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
