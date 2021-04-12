from sqlalchemy import ForeignKey, Enum, Column, Integer, String, DateTime, Boolean
from sqlalchemy.sql import func

from selections import db

gender_enum = Enum('Male', 'Female', 'Other', name='gender_enum')
interview_enum = Enum('Paper', 'Phone', name='interview_enum')


class Applicant(db.Model):
    __tablename__ = 'application'
    id = Column(Integer, primary_key=True)
    created = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    body = Column(String(6000), nullable=False)
    team = Column(Integer, nullable=False)
    gender = Column(gender_enum, nullable=False)
    phone_int = Column(Boolean, server_default='0', nullable=False)
    rit_id = Column(String(20), nullable=False)



class Members(db.Model):
    username = Column(String(50), primary_key=True)
    team = Column(Integer)


class Submission(db.Model):
    id = Column(Integer, primary_key=True)
    created = Column(DateTime(timezone=True), server_default=func.now(), nullable=False)
    application = Column(Integer, ForeignKey('application.id'), nullable=False)
    member = Column(String(50), ForeignKey('members.username'), nullable=False)
    medium = Column(interview_enum, primary_key=True)
    score = Column(Integer, nullable=False)


class Criteria(db.Model):
    id = Column(Integer, primary_key=True)
    name = Column(String(25), nullable=False)
    description = Column(String(100))
    min_score = Column(Integer, nullable=False)
    max_score = Column(Integer, nullable=False)
    weight = Column(Integer, nullable=False)
    medium = Column(interview_enum, nullable=False)
