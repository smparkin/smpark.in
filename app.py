from flask import Flask, render_template, request, Response
from flask_api import status
import sys

app = Flask(__name__)

'''with open(sys.path[0]+"/secrets", "r") as f:
    for i, line in enumerate(f):
        if i == 2:
            passwd = line[:-1]'''

# main smpark.in stuff
@app.route('/', methods=["GET"])
def home():
    return render_template('home.html')

@app.route('/welcome', methods=["GET"])
def welcome():
    return render_template('welcome.html')

@app.route('/privacy', methods=["GET"])
def privacy():
    return render_template('privacy.html')

@app.errorhandler(404)
def fourohfour(e):
    return render_template('404.html')

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000)