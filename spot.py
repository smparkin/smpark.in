import argparse
import requests
import json as jsn
import urllib.request
import sys
import time
import traceback

'''
create file called secrets in same folder as spot.py with app token on line 1 and refresh token on line 2
'''

def spotAuth():
    f = open(sys.path[0]+"/secrets", "r")
    appToken = f.readline()[:-1]
    refreshToken = f.readline()[:-1]
    tokenURL = "https://accounts.spotify.com/api/token"

    headers = {"Authorization": "Basic "+appToken}
    payload = {"grant_type": "refresh_token", "refresh_token": refreshToken}
    r = requests.post(tokenURL, headers=headers, data=payload)

    json = r.json()
    accessToken = json["access_token"]
    return accessToken


def spotDevice(headers, caller):
    r = requests.get("https://api.spotify.com/v1/me/player/devices", headers=headers)
    if r.status_code != 200:
        print('Invalid permissions!')
        quit()
    json = r.json()
    deviceid = None
    if len(json["devices"]) == 0:
        print("No playback devices")
        quit()
    elif len(json["devices"]) == 1:
        deviceid = json["devices"][0]["id"]
        devicename = json["devices"][0]["name"]
        if caller == "dev":
            print("\033[1m\033[92m"+devicename+"\033[0m is only device.")
            quit()
    elif caller == "vol" or caller == "prev" or caller == "next" or caller == "np" or caller == "play":
        for i in json["devices"]:
            if i["is_active"] == True:
                deviceid = i["id"]
                devicename = i["name"]
        if deviceid == None:
            print("No active device, defaulting to 0")
            deviceid = json["devices"][0]["id"]
            devicename = json["devices"][0]["name"]
    elif caller == "dev":
        j = 0
        devicedict = {}
        for i in json["devices"]:
            if i["is_active"] == False:
                print("["+str(j)+"] "+i["name"])
                devicedict.update( {j: [i["name"], i["id"]]})
                j += 1
        choice = input("Choose device: ")
        try:
            choice = int(choice)
        except:
            quit()
        deviceid = devicedict[choice][1]
        devicename = devicedict[choice][0]
    else:
        for i in range(0,len(json["devices"])):
            print("["+str(i)+"] "+json["devices"][i]["name"])
        choice = input("Choose device: ")
        try:
            choice = int(choice)
        except:
            quit()
        deviceid = json["devices"][choice]["id"]
        devicename = json["devices"][choice]["name"]
    
    devicedict = {}
    devicedict.update( {"deviceid": deviceid})
    devicedict.update( {"devicename": devicename})
    return devicedict


def spotSK(seekTime):
    accessToken = spotAuth()
    headers = {"Authorization": "Bearer "+accessToken}

    r = requests.get("https://api.spotify.com/v1/me/player/currently-playing", headers=headers)
    if r.status_code == 204:
        print("Nothing playing")
        quit()
    elif r.status_code != 200:
        print("Error: HTTP"+str(r.status_code))
        quit()
    json = r.json()
    durationMS = int(json["item"]["duration_ms"])
    seekMS = durationMS*(seekTime/4)
    seekMS = round(seekMS)

    r = requests.put("https://api.spotify.com/v1/me/player/seek?position_ms="+str(seekMS), headers=headers)
    if r.status_code == 204:
        print("Seeking to "+str((seekMS/1000)/60)+" minutes")
    else:
        print("Error: HTTP"+str(r.status_code))
    return r.status_code


def spotNP(imgcatBool, timeBool):
    accessToken = spotAuth()
    headers = {"Authorization": "Bearer "+accessToken}

    dev = spotDevice(headers, "np")
    devicename = dev["devicename"]

    r = requests.get("https://api.spotify.com/v1/me/player/currently-playing", headers=headers)
    if r.status_code == 204:
        print("Nothing playing")
        quit()
    elif r.status_code != 200:
        print("Error: HTTP"+str(r.status_code))
        quit()
    try:
        json = r.json()
        playing = json["is_playing"]
    except:
        playing = False

    if playing == False:
        text = "paused"
    elif playing == True:
        text = "playing" 

    title = json["item"]["name"]
    artist = json["item"]["album"]["artists"][0]["name"]
    imgurl = json["item"]["album"]["images"][1]["url"]
    durationSec = (json["item"]["duration_ms"])/1000
    currentSec = (json["progress_ms"])/1000
    print("\033[95m\033[1m"+title+"\033[0m by ", end='')
    print("\033[94m\033[1m"+artist+"\033[0m is "+text+" on ", end='')
    print("\033[92m\033[1m"+devicename+"\033[0m")
    if (timeBool):
        print(str(round(currentSec//60))+":"+(str(round(currentSec%60))).zfill(2)+"/"+str(round(durationSec//60))+":"+(str(round(durationSec%60)).zfill(2)))
    return r.status_code


def spotSE(context, query):
    print(query)
    accessToken = spotAuth()
    headers = {"Authorization": "Bearer "+accessToken}
    payload = {'type': context, 'q': query}

    r = requests.get("https://api.spotify.com/v1/search", params=payload, headers=headers)
    if r.status_code == 204:
        print("No results")
        quit()
    elif r.status_code != 200:
        print("Error: HTTP"+str(r.status_code))
        quit()
    json = r.json()
    try:
        if context == "album":
            uri = json["albums"]["items"][0]["uri"]
            name = json["albums"]["items"][0]["name"]
        elif context == "track":
            uri = json["tracks"]["items"][0]["uri"]
            name = json["tracks"]["items"][0]["name"]
    except:
        print("Search returned no results")
        quit()

    dev = spotDevice(headers, "search")

    if context == "album":
        payload = {"context_uri": uri}
    else:
        payload = {"uris": [uri]}
    r = requests.put("https://api.spotify.com/v1/me/player/play?device_id="+dev["deviceid"], headers=headers, data=jsn.dumps(payload))
    if r.status_code == 204:
        print("Playing "+context+"\033[1m\033[95m "+name+"\033[0m on \033[1m\033[92m"+dev["devicename"]+"\033[0m.")
    else:
        print("Unable to play \033[1m\033[95m"+name+"\033[0m.")
    return r.status_code


def spotSF():
    accessToken = spotAuth()
    headers = {"Authorization": "Bearer "+accessToken}
    r = requests.get("https://api.spotify.com/v1/me/player/currently-playing", headers=headers)
    if r.status_code == 204:
        print("No active playback session")
        quit()
    elif r.status_code != 200:
        print("Error: HTTP"+str(r.status_code))
        quit()

    r = requests.get("https://api.spotify.com/v1/me/player", headers=headers)
    json = r.json()
    shuf = json["shuffle_state"]
    if shuf == True:
        shuf = "false"
    elif shuf == False:
        shuf = "true"

    r = requests.put("https://api.spotify.com/v1/me/player/shuffle?state="+shuf, headers=headers)
    return r.status_code

def spotRE():
    accessToken = spotAuth()
    headers = {"Authorization": "Bearer "+accessToken}
    r = requests.get("https://api.spotify.com/v1/me/player/currently-playing", headers=headers)
    if r.status_code == 204:
        print("No active playback session")
        quit()
    elif r.status_code != 200:
        print("Error: HTTP"+str(r.status_code))
        quit()

    r = requests.get("https://api.spotify.com/v1/me/player", headers=headers)
    json = r.json()
    rep = json["repeat_state"]
    if rep == "context":
        rep = "off"
    elif rep == "off":
        rep = "context"
    else:
        rep = "off"

    r = requests.put("https://api.spotify.com/v1/me/player/repeat?state="+rep, headers=headers)
    if r.status_code == 204:
        r = requests.get("https://api.spotify.com/v1/me/player", headers=headers)
        json = r.json()
        currep = json["repeat_state"]
    else:
        print("Unable to toggle repeat.")
    return r.status_code


def spotPR():
    accessToken = spotAuth()
    headers = {"Authorization": "Bearer "+accessToken}
    r = requests.get("https://api.spotify.com/v1/me/player/currently-playing", headers=headers)
    if r.status_code == 204:
        print("No active playback session")
        quit()
    elif r.status_code != 200:
        print("Error: HTTP"+str(r.status_code))
        quit()
    json = r.json()
    trackname = json["item"]["name"]
    trackid = json["item"]["id"]

    dev = spotDevice(headers, "prev")

    r = requests.post("https://api.spotify.com/v1/me/player/previous", headers=headers)
    if r.status_code == 204:
        time.sleep(0.5)
        r = requests.get("https://api.spotify.com/v1/me/player/currently-playing", headers=headers)
        if r.status_code == 204:
            print("No active playback session")
            quit()
        elif r.status_code != 200:
            print("Error: HTTP"+str(r.status_code))
            quit()
        json = r.json()
        trackname = json["item"]["name"]
        trackid = json["item"]["id"]
    else:
        print("Unable to play \033[1m\033[95m"+trackname+"\033[0m.")
    return r.status_code


def spotNE():
    accessToken = spotAuth()
    headers = {"Authorization": "Bearer "+accessToken}
    r = requests.get("https://api.spotify.com/v1/me/player/currently-playing", headers=headers)
    if r.status_code == 204:
        print("No active playback session")
        quit()
    elif r.status_code != 200:
        print("Error: HTTP"+str(r.status_code))
        quit()
    json = r.json()
    trackname = json["item"]["name"]
    trackid = json["item"]["id"]

    dev = spotDevice(headers, "next")

    r = requests.post("https://api.spotify.com/v1/me/player/next", headers=headers)
    if r.status_code == 204:
        time.sleep(0.5)
        r = requests.get("https://api.spotify.com/v1/me/player/currently-playing", headers=headers)
        if r.status_code == 204:
            print("No active playback session")
            quit()
        elif r.status_code != 200:
            print("Error: HTTP"+str(r.status_code))
            quit()
        json = r.json()
        trackname = json["item"]["name"]
        trackid = json["item"]["id"]
    else:
        print("Unable to play \033[1m\033[95m"+trackname+"\033[0m.")
    return r.status_code


def spotPP():
    accessToken = spotAuth()
    headers = {"Authorization": "Bearer "+accessToken}

    dev = spotDevice(headers, "play")

    r = requests.get("https://api.spotify.com/v1/me/player", headers=headers)
    try:
        json = r.json()
        playing = json["is_playing"]
    except:
        playing = False

    if playing == False:
        r = requests.put("https://api.spotify.com/v1/me/player/play?device_id="+dev["deviceid"], headers=headers)
        if r.status_code == 204:
            time.sleep(0.5)
            r = requests.get("https://api.spotify.com/v1/me/player/currently-playing", headers=headers)
            if r.status_code == 204:
                quit()
            elif r.status_code != 200:
                print("Error: HTTP"+str(r.status_code))
                quit()
            json = r.json()
            trackname = json["item"]["name"]
            trackid = json["item"]["id"]
    elif playing == True:
        r = requests.put("https://api.spotify.com/v1/me/player/pause?device_id="+dev["deviceid"], headers=headers)
        if r.status_code == 204:
            r = requests.get("https://api.spotify.com/v1/me/player/currently-playing", headers=headers)
            if r.status_code == 204:
                quit()
            elif r.status_code != 200:
                quit()
            json = r.json()
            trackname = json["item"]["name"]
            trackid = json["item"]["id"]
    return "Success"


def spotLS():
    accessToken = spotAuth()
    headers = {"Authorization": "Bearer "+accessToken}
    r = requests.get("https://api.spotify.com/v1/me/player/currently-playing", headers=headers)
    if r.status_code == 204:
        print("Nothing playing")
        quit()
    elif r.status_code != 200:
        print("Error: HTTP"+str(r.status_code))
        quit()
    json = r.json()
    trackname = json["item"]["name"]
    trackid = json["item"]["id"]

    headers = {"Authorization": "Bearer "+accessToken, "Accept": "application/json", "Content-Type": "application/json"}
    r = requests.put("https://api.spotify.com/v1/me/tracks?ids="+trackid, headers=headers)
    if r.status_code == 200:
        print("Added \033[1m\033[95m"+json["item"]["name"]+"\033[0m to Liked Songs")
    else:
        print("An error occured, fun")
    return r.status_code

def spotRL():
    accessToken = spotAuth()
    headers = {"Authorization": "Bearer "+accessToken}
    r = requests.get("https://api.spotify.com/v1/me/player/currently-playing", headers=headers)
    if r.status_code == 204:
        print("Nothing playing")
        quit()
    elif r.status_code != 200:
        print("Error: HTTP"+str(r.status_code))
        quit()
    json = r.json()
    trackname = json["item"]["name"]
    trackid = json["item"]["id"]

    headers = {"Authorization": "Bearer "+accessToken, "Accept": "application/json", "Content-Type": "application/json"}
    r = requests.delete("https://api.spotify.com/v1/me/tracks?ids="+trackid, headers=headers)
    if r.status_code == 200:
        print("Removed \033[1m\033[95m"+json["item"]["name"]+"\033[0m from Liked Songs")
    else:
        print("An error occured, fun")


def spotAP():
    accessToken = spotAuth()
    headers = {"Authorization": "Bearer "+accessToken}
    r = requests.get("https://api.spotify.com/v1/me/player/currently-playing", headers=headers)
    if r.status_code == 204:
        print("Nothing playing")
        quit()
    elif r.status_code != 200:
        print("Error: HTTP"+str(r.status_code))
        quit()
    json = r.json()
    print("Add \033[1m\033[95m"+json["item"]["name"]+"\033[0m to:")
    trackname = json["item"]["name"]
    trackuri = json["item"]["uri"]
    trackuri.replace(":", "%3A")

    r = requests.get("https://api.spotify.com/v1/me", headers=headers)
    json = r.json()
    userid = json["id"]

    r = requests.get("https://api.spotify.com/v1/me/playlists?limit=50", headers=headers)
    json = r.json()
    j = 0
    playdict = {}
    for i in json["items"]:
        if userid == i["owner"]["display_name"]:
            print("["+str(j)+"] "+i["name"])
            playdict.update( {j: [i["name"], i["id"]]})
            j += 1
        elif i["collaborative"] == True:
            print("["+str(j)+"] "+i["name"])
            playdict.update( {j: [i["name"], i["id"]]})
            j += 1
    choice = input("Select Playlist: ")
    try:
        choice = int(choice)
    except:
        quit()
    playlistid = playdict[choice][1]
    playlistname = playdict[choice][0]

    headers = {"Authorization": "Bearer "+accessToken, "Accept": "application/json", "Content-Type": "application/json"}
    r = requests.post("https://api.spotify.com/v1/playlists/"+playlistid+"/tracks?uris="+trackuri, headers=headers)
    if r.status_code == 201:
        print("Successfully added \033[1m\033[95m"+trackname+"\033[0m to \033[1m\033[96m"+playlistname+"\033[0m")
    else:
        print("Unable to add song to specified playlist. Do you have access to do so?")
    return r.status_code


def spotPD():
    accessToken = spotAuth()
    headers = {"Authorization": "Bearer "+accessToken}

    dev = spotDevice(headers, "dev")

    payload = {"device_ids":[dev["deviceid"]]}
    r = requests.put("https://api.spotify.com/v1/me/player", headers=headers, data=jsn.dumps(payload))
    if r.status_code == 204:
        time.sleep(0.5)
        r = requests.get("https://api.spotify.com/v1/me/player/currently-playing", headers=headers)
        if r.status_code == 204:
            print("No active playback session")
            quit()
        elif r.status_code != 200:
            print("Error: HTTP"+str(r.status_code))
            quit()
        json = r.json()
        trackname = json["item"]["name"]
        trackid = json["item"]["id"]
        print("Playing \033[1m\033[95m"+trackname+"\033[0m on \033[1m\033[92m"+dev["devicename"]+"\033[0m.")
    elif r.status_code == 202:
        time.sleep(2)
        r = requests.get("https://api.spotify.com/v1/me/player/currently-playing", headers=headers)
        if r.status_code == 204:
            print("No active playback session")
            quit()
        elif r.status_code != 200:
            print("Error: HTTP"+str(r.status_code))
            quit()
        json = r.json()
        trackname = json["item"]["name"]
        trackid = json["item"]["id"]
        print("Playing \033[1m\033[95m"+trackname+"\033[0m on \033[1m\033[92m"+dev["devicename"]+"\033[0m.")
    else:
        print(r.status_code)
        print("Unable to transfer playback.")

   
def spotRP():
    accessToken = spotAuth()
    headers = {"Authorization": "Bearer "+accessToken}
    r = requests.get("https://api.spotify.com/v1/me/player/currently-playing", headers=headers)
    if r.status_code == 204:
        print("Nothing playing")
        quit()
    elif r.status_code != 200:
        print("Error: HTTP"+str(r.status_code))
        quit()
    json = r.json()
    print("Remove \033[1m\033[95m"+json["item"]["name"]+"\033[0m from:")
    trackname = json["item"]["name"]
    trackuri = json["item"]["uri"]

    r = requests.get("https://api.spotify.com/v1/me", headers=headers)
    json = r.json()
    userid = json["id"]

    r = requests.get("https://api.spotify.com/v1/me/playlists?limit=50", headers=headers)
    json = r.json()
    j = 0
    playdict = {}
    for i in json["items"]:
        if userid == i["owner"]["display_name"]:
            print("["+str(j)+"] "+i["name"])
            playdict.update( {j: [i["name"], i["id"]]})
            j += 1
        elif i["collaborative"] == True:
            print("["+str(j)+"] "+i["name"])
            playdict.update( {j: [i["name"], i["id"]]})
            j += 1
    choice = input("Select Playlist: ")
    try:
        choice = int(choice)
    except:
        quit()
    playlistid = playdict[choice][1]
    playlistname = playdict[choice][0]

    headers = {"Authorization": "Bearer "+accessToken, "Accept": "application/json", "Content-Type": "application/json"}
    payload = { "tracks": [{ "uri": trackuri }] }
    r = requests.delete("https://api.spotify.com/v1/playlists/"+playlistid+"/tracks", headers=headers, data=jsn.dumps(payload))
    if r.status_code == 200:
        print("If \033[1m\033[95m"+trackname+"\033[0m was in \033[1m\033[96m"+playlistname+"\033[0m it has been removed.")
    else:
        print("Unable to remove song from specified playlist. Do you have access to do so?")
    return r.status_code


def spotPL():
    accessToken = spotAuth()
    headers = {"Authorization": "Bearer "+accessToken}
    print("Play:")

    r = requests.get("https://api.spotify.com/v1/me/playlists?limit=50", headers=headers)
    json = r.json()
    j = 0
    playdict = {}
    for i in json["items"]:
        print("["+str(j)+"] "+i["name"])
        playdict.update( {j: [i["name"], i["id"]]})
        j += 1
    choice = input("Select Playlist: ")
    try:
        choice = int(choice)
    except:
        quit()
    playlistid = playdict[choice][1]
    playlistname = playdict[choice][0]

    dev = spotDevice(headers, "playlist play")

    payload = {"context_uri": "spotify:playlist:"+playlistid}
    r = requests.put("https://api.spotify.com/v1/me/player/play?device_id="+dev["deviceid"], headers=headers, data=jsn.dumps(payload))
    if r.status_code == 204:
        time.sleep(0.5)
        r = requests.get("https://api.spotify.com/v1/me/player/currently-playing", headers=headers)
        if r.status_code == 204:
            print("No active playback session")
            quit()
        elif r.status_code != 200:
            print("Error: HTTP"+str(r.status_code))
            quit()
        json = r.json()
        trackname = json["item"]["name"]
        trackid = json["item"]["id"]
        print("Playing \033[1m\033[95m"+trackname+"\033[0m on \033[1m\033[92m"+dev["devicename"]+"\033[0m.")
    else:
        print("Unable to play \033[1m\033[95m"+trackname+"\033[0m.")
    return r.status_code


def spotVL(vol):
    accessToken = spotAuth()
    headers = {"Authorization": "Bearer "+accessToken}

    dev = spotDevice(headers, "vol")

    r = requests.get("https://api.spotify.com/v1/me/player", headers=headers)
    print(r)
    json = r.json()
    curVol = json["device"]["volume_percent"]
    curVol = int(curVol)

    if vol == "up":
        vol = (curVol+10)
        if vol > 100:
            vol = 100
    elif vol == "down":
        vol = (curVol-10)
        if vol < 0:
            vol = 0

    r = requests.put("https://api.spotify.com/v1/me/player/volume?volume_percent="+str(vol), headers=headers)
    if r.status_code == 204:
        print("Volume on \033[1m\033[92m"+dev["devicename"]+"\033[0m set to "+str(vol))
    else:
        json = r.json()
        reason = json["error"]["reason"]
        if reason == "VOLUME_CONTROL_DISALLOW":
            print("Device \033[1m\033[92m"+dev["devicename"]+"\033[0m does not allow volume to be controlled through API")
        else:
            print("No active playback devices")
    return r.status_code
