import collections, glob, json, statistics, sys

# go-json against sonic per category, payload and condition, for master and head: the median of the rounds.
res = collections.defaultdict(list)
for path in glob.glob(sys.argv[1] + '/*-*.json'):
    ref = path.split('/')[-1].split('-')[0]
    for r in json.load(open(path))['results']:
        res[(ref, r['condition'], r['payload'], r['config'])].append(r['nsPerOp'])
med = {k: statistics.median(v) for k, v in res.items()}
for cond in ['live-heap', 'no-live-heap']:
    print('==', cond)
    print(f"{'category':8} {'payload':12} {'master':>18} {'head':>18} {'head/master':>12}")
    for cat in ['std', 'fast', 'v2', 'fastest']:
        for p in ['small', 'medium', 'large', 'twitter', 'twitter-any', 'github', 'openai', 'anthropic', 'code']:
            row = f"{cat:8} {p:12}"
            vals = {}
            for ref in ['master', 'head']:
                g = med.get((ref, cond, p, cat + '/go-json'))
                s = med.get((ref, cond, p, cat + '/sonic'))
                vals[ref] = g
                row += f" {g:8.0f} {(g / s - 1) * 100:+7.1f}%" if g and s else f" {'-':>18}"
            if vals.get('master') and vals.get('head'):
                row += f" {(vals['head'] / vals['master'] - 1) * 100:+11.1f}%"
            print(row)
