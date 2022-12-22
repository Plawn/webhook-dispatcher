import requests


def test():
    r = requests.post('http://localhost:8000/send', json={"payload": "dev", "addresses":["https://yay"]})
    print(r)


if __name__ == "__main__":
    test()