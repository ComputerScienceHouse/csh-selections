from sqlalchemy import ForeignKey, Enum, Column, Integer, String

from selections import db

gender_enum = Enum('Male', 'Female', 'Other', name='gender_enum')
interview_enum = Enum('Paper', 'Phone', name='interview_enum')


class applicant(db.Model):
    __tablename__ = "application"
    id = Column(Integer, primary_key = True)
    body = Column(String(5000))
    team = Column(Integer)
    gender = Column(gender_enum)


class members(db.Model):
    username = Column(String(50), primary_key = True)
    team = Column(Integer)


class submission(db.Model):
    id = Column(Integer, primary_key = True)
    application = Column(Integer, ForeignKey("application.id"))
    member = Column(String(50), ForeignKey("members.username"))
    medium = Column(interview_enum)
    score = Column(Integer)


class criteria(db.Model):
	id = Column(Integer, primary_key = True)
	name = Column(String(25))
	description = Column(String(100))
	min_score = Column(Integer)
	max_score = Column(Integer)
	weight = Column(Integer)
	medium = Column(interview_enum)
