"""
Helper to render dependencies.
"""

import collections
import sys
import textwrap


def dump_deps(task=None, indent=0, dot=False, file=file):
    """
    Print dependency graph for a given task to stdout.
    """
    if dot is True:
        return dump_deps_dot(task, file=file)
    if task is None:
        return
    g = build_dep_graph(task)
    print('%s \_ %s' % ('   ' * indent, task), file=file)
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
    print(textwrap.dedent("""
          graph [fontname=helvetica];
          node [shape=record fontname=helvetica];
    """),
          file=file)
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
