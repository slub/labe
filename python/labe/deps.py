import collections
import textwrap


def dump_deps(task=None, indent=0, dot=False):
    """
    Print dependency graph for a given task to stdout.
    """
    if dot is True:
        return dump_deps_dot(task)
    if task is None:
        return
    g = build_dep_graph(task)
    print('%s \_ %s' % ('   ' * indent, task))
    for dep in g[task]:
        dump_deps(task=dep, indent=indent + 1)


def dump_deps_dot(task=None):
    if task is None:
        return
    g = build_dep_graph(task)
    print("digraph deps {")
    print(
        textwrap.dedent("""
          graph [fontname=helvetica];
          node [shape=record fontname=helvetica];
    """))
    for k, vs in g.items():
        for v in vs:
            print(""" "{}" -> "{}"; """.format(k, v))
    print("}")


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
