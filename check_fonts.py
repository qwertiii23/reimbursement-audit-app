import matplotlib.font_manager as fm
fonts = set(f.name for f in fm.fontManager.ttflist)
for f in sorted(fonts):
    if any(k in f.lower() for k in ['sf', 'helv', 'arial', 'ping', 'mono', 'menlo', 'courier', 'sans']):
        print(f)
