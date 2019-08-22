from flask import Flask, json
import requests
from statsmodels.tsa.arima_model import ARIMA
import time


cpu_scale = {"scale": 1}
memory_scale = {"scale": 1}
# url = 'http://prometheus-system-discovery.knative-monitoring.svc.cluster.local:9090'
url = 'http://localhost:9090'
api = Flask(__name__)

@api.route('/cpu', methods=['GET'])
def get_cpu_scale():
  """
  Forcast pods using the cpu model
  """
  params = 'avg (rate (container_cpu_usage_seconds_total{image!="",namespace="default",container_name="user-container"}[1m]))'

  end = time.time()
  start = end - 60*60

  try:
    resp = requests.get('{0}/api/v1/query_range'.format(url),
        params={'query': params, 
                'start':start,
                'end':end,
                'step':14})
    print(resp.content)
    results = resp.json()['data']['result']
    if resp.status_code != 200:
        # This means something went wrong.
        print("xxx")
    else:
        print("succeeded...")
        print(results)
        for result in results:
            for val in result['values']:
                print(val[0], val[1])
            print(result['metric'].keys())

    # on error send the last value
    return json.dumps(cpu_scale)
  except requests.ConnectionError as ex:
    print("Prometheus Connection Error cpu...: ({0})".format(url))
    return json.dumps({"state": "failed"})

@api.route('/memory', methods=['GET'])
def get_memory_scale():
    params='container_memory_usage_bytes{namespace="default",container_name="user-container"}'

    try:
        resp = requests.get('{0}/api/v1/query'.format(url),
            params={'query': params})
        results = resp.json()['data']['result']
        if resp.status_code != 200:
            # This means something went wrong.
            print("xxx")
        else:
            print("succeeded...")
            print(results)
            for result in results:
                print(result['metric'].keys())

        return json.dumps(memory_scale)
    except requests.ConnectionError as ex:
        print("Prometheus Connection Error memory...")
        return json.dumps({"state": "failed"})



if __name__ == '__main__':
    api.run()
    # get_cpu_scale()

