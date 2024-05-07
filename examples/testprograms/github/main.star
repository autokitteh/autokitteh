load("@github", "mygithub1")

def on_github_issue_comment(data):
    print(data)
    print(mygithub1.get_issue(repo=data.repo.name, owner=data.repo.owner.login, number=data.issue.number))
    
