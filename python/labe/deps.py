"""
Helper to render dependencies.

Example output:

  $ labe.pyz --deps CombinedUpdate
   \_ CombinedUpdate(date=2022-01-21)
      \_ IdMappingDatabase(date=2022-01-21)
         \_ IdMappingTable(date=2022-01-21)
            \_ SolrFetchDocs(date=2022-01-21, name=ai, short=True)
            \_ SolrFetchDocs(date=2022-01-21, name=slub-production, short=False)
            \_ SolrFetchDocs(date=2022-01-21, name=main, short=False)
      \_ OpenCitationsDatabase()
         \_ OpenCitationsSingleFile()
            \_ OpenCitationsDownload()
      \_ SolrDatabase(date=2022-01-21, name=ai, short=True)
         \_ SolrFetchDocs(date=2022-01-21, name=ai, short=True)
      \_ SolrDatabase(date=2022-01-21, name=slub-production, short=False)
         \_ SolrFetchDocs(date=2022-01-21, name=slub-production, short=False)
      \_ SolrDatabase(date=2022-01-21, name=main, short=True)
         \_ SolrFetchDocs(date=2022-01-21, name=main, short=True)

"""

import collections
import os
import sys
import textwrap


def dump_deps(task=None, indent=0, dot=False, file=sys.stdout):
    """
    Print dependency graph for a given task to stdout.
    """
    if dot is True:
        return dump_deps_dot(task, file=file)
    if task is None:
        return
    g = build_dep_graph(task)
    mark = "\033[31m𐄂\033[0m"
    if task.output() and os.path.exists(task.output().path):
        mark = "\033[32m√\033[0m"
    print("{} \_ {} {}".format("   " * indent, mark, task), file=file)
    for dep in g[task]:
        dump_deps(task=dep, indent=indent + 1)


def dump_deps_dot(task=None, file=sys.stdout):
    """
    Render a graphviz dot representation.
    """
    if task is None:
        return
    g = build_dep_graph(task)
    print("digraph deps {", file=file)
    print(
        textwrap.dedent("""
          graph [fontname=helvetica];
          node [shape=record fontname=helvetica];
    """),
        file=file,
    )
    for k, vs in g.items():
        for v in vs:
            print(""" "{}" -> "{}"; """.format(k, v), file=file)
    print("}", file=file)


def build_dep_graph(task=None):
    """
    Return the task graph as dict mapping nodes to children.
    """
    g = collections.defaultdict(set)
    queue = [task]
    while len(queue) > 0:
        task = queue.pop()
        for dep in task.deps():
            g[task].add(dep)
            queue.append(dep)
    return g
