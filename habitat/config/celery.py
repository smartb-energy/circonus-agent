#!{{pkgPathFor "core/python"}}/bin/python

import urllib.request
import json

flower_host="http://localhost:5555"

resp = urllib.request.urlopen("{}/api/workers".format(flower_host)).read()
data = json.loads(resp)
registered = []
for worker in data.values():
    if "registered" in worker:
      registered += worker["registered"]
registered = set(registered)

print("{}\t{}\t{}".format("celery_task_count", "L", len(registered)))

for task in registered:
    task_response = urllib.request.urlopen("{}/api/tasks?taskname={}".format(flower_host, task)).read()
    task_data = json.loads(task_response)
    count = len([task_name for task_name in task_data.values() if task_name["state"] == "SUCCESS"])
    print("{}\t{}\t{}".format(task, "L", count))
