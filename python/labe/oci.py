# -*- coding: utf-8 -*-
"""
Access the open citations download site and dataset.

    >>> from labe.oci import OpenCitationsDataset
    >>> ds = OpenCitationsDataset()
    >>> ds.most_recent_download_url(format="CSV")
    'https://figshare.com/ndownloader/articles/6741422/versions/11'

"""

import re

import requests

__all__ = [
    'OpenCitationsDataset',
]

default_user_agent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/93.0.4577.63 Safari/537.36"


def get_terminal_url(link, headers=None, user_agent=default_user_agent):
    """
    Given a url, return the terminal url.
    """
    if headers is None:
        headers = {
            "User-Agent": user_agent,
        }
    resp = requests.get(link, headers=headers)
    if resp.status_code >= 400:
        raise RuntimeError("got http status {} on {}".format(
            resp.status_code, link))
    if not resp.url:
        raise ValueError('url not found')
    return resp.url


def get_figshare_download_link(link):
    """
    Given a link that should redirect to figshare. If it does not, this will fail.
    """
    landing_page_url = get_terminal_url(link)
    pattern_figshare_url = re.compile(
        r"https://figshare.com/articles/dataset/"
        r"(?P<name>[^/]*)/(?P<id>[^/]*)/(?P<version>[\d]*)")
    match = re.match(pattern_figshare_url, landing_page_url)
    if not match:
        raise RuntimeError(
            "unexpected landing page url: {}".format(landing_page_url))
    groups = match.groupdict()
    return "https://figshare.com/ndownloader/articles/{}/versions/{}".format(
        groups["id"], groups["version"])


class OpenCitationsDataset:
    """
    Encapsulates dataset download. We have a few scrapes and hops, e.g.

    https://opencitations.net/download
                    |
                    v
    https://doi.org/10.6084/m9.figshare.6741422.v11
                    |
                    v
    https://figshare.com/articles/dataset/Crossref_Open_Citation_Index_CSV_dataset_of_all_the_citation_data/6741422/11
                    |
                    v
    https://figshare.com/ndownloader/articles/6741422/versions/11

    This is fragile, since OCI and figshare both may change their scheme.
    Figshare has an API, https://docs.figshare.com, but the complete download
    URL is not contained in the articles details, e.g. in
    https://api.figshare.com/v2/articles/6741422.

    Use `direct_download_url` to override link returned by
    `most_recent_download_url`, e.g. if scraping breaks.
    """
    def __init__(self, direct_download_url=None):
        self.download_url = "https://opencitations.net/download"
        self.cache = dict()
        resp = requests.get(self.download_url)
        if resp.status_code >= 400:
            raise RuntimeError("cannot download page at: {}".format(
                self.download_url))
        self.cache[self.download_url] = resp.text
        self.direct_download_url = direct_download_url

    def __repr__(self):
        return self.__str__()

    def __str__(self):
        return "<OpenCitationsDataset via {} ({} rows)>".format(
            self.download_url, len(self.rows()))

    def rows(self):
        """
        Citation data (CSV)</td><td><a
        href="https://doi.org/10.6084/m9.figshare.6741422.v11">ZIP</a></td><td>193.6
        GB (29.44 GB zipped)

        Return list of dicts describing a table entry. Most recent should be the first.

            [ ...
	        {'format': 'CSV',
	         'url': 'https://doi.org/10.6084/m9.figshare.6741422.v3',
	         'ext': 'ZIP',
	         'size': '72 GB ',
	         'size_compressed': '11 GB zipped'},
	        {'format': 'N-Triple',
	         'url': 'https://doi.org/10.6084/m9.figshare.6741425.v3',
	         'ext': 'ZIP',
	         'size': '481 GB ',
	         'size_compressed': '22 GB zipped'},
             ... ]

        """
        pattern_citation_data = re.compile(
            r'citation data \((?P<format>[^)]*)\)</td>'
            r'<td><a href="(?P<url>[^"]*)">(?P<ext>'
            r'[^<]*)</a></td><td>(?P<size>[^(]*)\((?P<'
            r'size_compressed>[^)]*)\)</td></tr>',
            re.IGNORECASE,
        )
        return [
            m.groupdict() for m in re.finditer(pattern_citation_data,
                                               self.cache[self.download_url])
        ]

    def most_recent_url(self, format="CSV"):
        """
        Return the most recent link for a given format (as it is written on the
        website, e.g. "CSV").
        """
        rows = [row for row in self.rows() if row.get("format") == format]
        if not rows:
            raise ValueError("no url found for format {}".format(format))
        return rows[0].get("url")

    def most_recent_download_url(self, format="CSV"):
        """
        Get the final download link, we want something like:
        https://figshare.com/ndownloader/articles/6741422/versions/11.
        """
        if self.direct_download_url:
            return self.direct_download_url
        return get_figshare_download_link(self.most_recent_url(format=format))
