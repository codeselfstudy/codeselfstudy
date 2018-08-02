from bs4 import BeautifulSoup


DATA_DIR = './data/'


def extract_content(fname):
    """Read in an HTML file and extract the problem title and body."""
    with open('{}{}.html'.format(DATA_DIR, fname), 'r') as f:
        raw_content = f.read()

    soup = BeautifulSoup(raw_content, 'html.parser')

    title = soup.find(id='content').find('h2').text
    body = soup.find('div', {'class': 'problem_content'})

    return {
        'title': title,
        'body': body,
    }
