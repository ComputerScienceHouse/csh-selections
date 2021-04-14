"""Add application rit_id

Revision ID: 65977197d0bd
Revises: f7fc75087f9f
Create Date: 2021-04-11 22:15:29.825856

"""
from alembic import op
import sqlalchemy as sa
from sqlalchemy import orm
from sqlalchemy.ext.declarative import declarative_base


# revision identifiers, used by Alembic.
revision = '65977197d0bd'
down_revision = 'f7fc75087f9f'
branch_labels = None
depends_on = None

Base = declarative_base()


class Applicant(Base):
    __tablename__ = 'application'
    id = sa.Column(sa.Integer, primary_key=True)
    rit_id = sa.Column(sa.String(20), nullable=False)


def upgrade():
    op.add_column('application', sa.Column('rit_id', sa.String(20), nullable=False))

    bind = op.get_bind()
    session = orm.Session(bind=bind)

    for application in session.query(Applicant):
        application.rit_id = str(application.id)

    session.commit()

def downgrade():
    op.drop_column('application', 'rit_id')
