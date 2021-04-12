from werkzeug.middleware.profiler import ProfilerMiddleware
from selections import app

app.config['PROFILE'] = True
app.wsgi_app = ProfilerMiddleware(app.wsgi_app, restrictions=[30], sort_by=('cumtime',))
app.run(debug = True)
