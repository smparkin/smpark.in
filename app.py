from flask import Flask, render_template, request, Response
from flask_api import status
from spot import *
import sys

app = Flask(__name__)

with open(sys.path[0]+"/secrets", "r") as f:
    for i, line in enumerate(f):
        if i == 2:
            passwd = line[:-1]

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

# ndsspotify stuff
@app.route('/play', methods=["POST"])
def flaskPlay():
    print(request.headers)
    if request.headers['Auth'] == passwd:
        spotPP()
        return Response('{"status":"Success"}', status=200, mimetype='application/json')
    else:
        return Response('{"status":"Incorrect password"}', status=401, mimetype='application/json')

@app.route('/next', methods=["POST"])
def flaskNext():
    if request.headers['Auth'] == passwd:
        spotNE()
        return Response('{"status":"Success"}', status=200, mimetype='application/json')
    else:
        return Response('{"status":"Incorrect password"}', status=401, mimetype='application/json')

@app.route('/prev', methods=["POST"])
def flaskPrev():
    if request.headers['Auth'] == passwd:
        spotPR()
        return Response('{"status":"Success"}', status=200, mimetype='application/json')
    else:
        return Response('{"status":"Incorrect password"}', status=401, mimetype='application/json')

@app.route('/shuf', methods=["POST"])
def flaskShuf():
    if request.headers['Auth'] == passwd:
        spotSF()
        return Response('{"status":"Success"}', status=200, mimetype='application/json')
    else:
        return Response('{"status":"Incorrect password"}', status=401, mimetype='application/json')

@app.route('/repr', methods=["POST"])
def flaskRepr():
    if request.headers['Auth'] == passwd:
        spotRE()
        return Response('{"status":"Success"}', status=200, mimetype='application/json')
    else:
        return Response('{"status":"Incorrect password"}', status=401, mimetype='application/json')

@app.route('/sear', methods=["POST"])
def flaskSear():
    if request.headers['Auth'] == passwd:
        spotSE("track", request.headers['Search-Text'])
        return Response('{"status":"Success"}', status=200, mimetype='application/json')
    else:
        return Response('{"status":"Incorrect password"}', status=401, mimetype='application/json')

@app.route('/like', methods=["POST"])
def flaskLike():
    if request.headers['Auth'] == passwd:
        spotLS()
        return Response('{"status":"Success"}', status=200, mimetype='application/json')
    else:
        return Response('{"status":"Incorrect password"}', status=401, mimetype='application/json')

@app.route('/disl', methods=['POST'])
def flaskDisl():
    if request.headers['Auth'] == passwd:
        spotRL()
        return Response('{"status":"Success"}', status=200, mimetype='application/json')
    else:
        return Response('{"status":"Incorrect password"}', status=401, mimetype='application/json')

@app.route('/volu', methods=["POST"])
def flaskVolu():
    if request.headers['Auth'] == passwd:
        spotVL(request.headers['Volume'])
        return Response('{"status":"Success"}', status=200, mimetype='application/json')
    else:
        return Response('{"status":"Incorrect password"}', status=401, mimetype='application/json')

if __name__ == '__main__':
    app.run(host='0.0.0.0', port=5000)