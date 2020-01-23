import json
from bs4 import BeautifulSoup
from os import listdir
from os.path import isfile, join


DATA_DIR = "../data/project_euler"


def extract_content(fname):
    """Read in an HTML file and extract the problem title and body."""
    with open("{}{}".format(DATA_DIR, fname), "r") as f:
        raw_content = f.read()

    soup = BeautifulSoup(raw_content, "html.parser")

    problem_number = fname.split(".")[0]
    title = soup.find(id="content").find("h2").text
    body = soup.find("div", {"class": "problem_content"})
    url = f"https://projecteuler.net/problem={problem_number}"

    return {
        "title": title,
        "body": str(body),
        "url": url,
    }


if __name__ == "__main__":
    files = [f for f in listdir(DATA_DIR) if isfile(join(DATA_DIR, f))]

    for file in files:
        output = extract_content(file)
        outfile_name = file.split(".")[0] + ".json"
        with open(outfile_name, "w") as f:
            f.write(json.dumps(output, indent=4))
