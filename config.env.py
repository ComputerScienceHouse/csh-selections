import os
import secrets
from os import environ as env

# Flask config
DEBUG = True
IP = env.get('IP', 'localhost')
PORT = env.get('PORT', 8080)
SERVER_NAME = env.get('SERVER_NAME', 'selections-dev.csh.rit.edu')

# DB Info
SQLALCHEMY_DATABASE_URI = env.get('SQLALCHEMY_DATABASE_URI', 'sqlite:///{}'.format(os.path.join(os.getcwd(), "data.db")))
SQLALCHEMY_TRACK_MODIFICATIONS = False
SQLALCHEMY_POOL_RECYCLE = 299
SQLALCHEMY_POOL_TIMEOUT = 20
# Openshift secret
SECRET_KEY = env.get("SECRET_KEY", default=''.join(secrets.token_hex(16)))

# OpenID Connect SSO config
OIDC_ISSUER = env.get('OIDC_ISSUER', 'https://sso.csh.rit.edu/auth/realms/csh')
OIDC_CLIENT_CONFIG = {
    'client_id': env.get('OIDC_CLIENT_ID', 'selections'),
    'client_secret': env.get('OIDC_CLIENT_SECRET', ''),
    'post_logout_redirect_uris': [env.get('OIDC_LOGOUT_REDIRECT_URI', 'https://profiles.csh.rit.edu/logout')]
}

LDAP_BIND_DN = env.get("LDAP_BIND_DN", default="cn=selections2,ou=Apps,dc=csh,dc=rit,dc=edu")
LDAP_BIND_PASS = env.get("LDAP_BIND_PASS", default=None)
