import time
# import json
# import logging
import requests
from random import uniform


URL_TEMPLATE = "https://projecteuler.net/problem={}"
DATA_DIR = "./data"


if __name__ == "__main__":
    print("running program")

    print("starting download loop")
    for n in range(1, 624):  # The Project Euler problems go from 1-623
        current_url = URL_TEMPLATE.format(n)
        print("fetching problem #{} at {}".format(n, current_url))

        r = requests.get(current_url)

        if r.status_code == 200:
            print("got status_code {} for problem #{}".format(r.status_code, n))
            with open("{}{}.html".format(DATA_DIR, n), "w") as f:
                print("writing file for problem #{}".format(n))
                f.write(r.text)
        else:
            print("couldn\"t download problem #{}".format(n))
        time.sleep(uniform(1, 4))
    print("done")
