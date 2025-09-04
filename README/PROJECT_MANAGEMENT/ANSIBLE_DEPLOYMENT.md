The following is a tentative plan for using ansible to deploy to Digital Ocean.
This includes a draft on the usage of Github pipelines.

Project Plan: SocialPredict Droplet Automation & Maintenance
Goal
Automate deployment and maintenance of the SocialPredict stack (from github.com/openpredictionmarkets/socialpredict) on a DigitalOcean droplet, ensuring:

Consistent repo cloning and branch management.

Docker & Docker Compose are always installed and configured.

Disk space issues (log bloat, dangling images) are automatically cleaned up.

Optionally, deployment is run automatically as part of the playbook.

Deliverables
An Ansible playbook (or role) to:

Clone/update the repo into /opt/socialpredict.

Ensure Docker and Docker Compose are installed and configured.

Apply Docker log rotation (/etc/docker/daemon.json).

Set up cron jobs for weekly pruning and daily log cleanup.

Optionally trigger a full ./SocialPredict up deployment.

Optional: A separate deployment playbook for pushing new branches or images (so maintenance and deployments can be run separately).

Documentation in deploy/README.md explaining:

How to set up the inventory file.

How to run the playbook(s).

How to trigger redeploys.

Phases
Phase 1: Environment Setup
 Create Ansible directory structure: deploy/ansible/.

 Add hosts.ini inventory for the droplet (IP + ansible_user=root).

 Ensure Ansible can connect (test ansible all -i hosts.ini -m ping).

Phase 2: Maintenance Playbook
 Playbook tasks:

Install docker.io, docker-compose, git.

Clone socialpredict into /opt/socialpredict (branch configurable).

Configure Docker log rotation (10MB × 3 files).

Add cron jobs:

Weekly docker system prune -af --volumes.

Daily truncate container logs.

Restart Docker if config changes.

Phase 3: Deployment (Optional, Separate or Combined)
 Build local images (socialpredict-backend, socialpredict-frontend) from specified branch.

 Run ./SocialPredict install and ./SocialPredict up.

 Tag and optionally push images to GHCR for future redeploys.

Phase 4: Testing
 Run playbook on the droplet.

 Verify:

Repo is in /opt/socialpredict.

Cron jobs exist (crontab -l).

Logs rotate and truncate correctly.

Docker prune runs weekly.

./SocialPredict up starts the stack cleanly.

Project Plan 1: Get Ansible Running for SocialPredict
Goal
Set up Ansible so we can manage the SocialPredict droplet from our local machine (or a CI/CD runner) and handle deployments systematically.

Steps
Local Control Node Setup

Install Ansible on your laptop or CI runner:

bash
Copy
Edit
brew install ansible    # on Mac
Verify version:

bash
Copy
Edit
ansible --version
Inventory File

Create deploy/ansible/hosts.ini:

ini
Copy
Edit
[socialpredict_droplet]
143.198.177.112 ansible_user=root
Test connectivity:

bash
Copy
Edit
ansible all -i deploy/ansible/hosts.ini -m ping
Basic Ansible Playbook

Make deploy/ansible/setup.yml:

yaml
Copy
Edit
---
- name: Initial SocialPredict setup
  hosts: socialpredict_droplet
  become: true
  tasks:
    - name: Ensure Git, Docker, and Compose are installed
      apt:
        name:
          - git
          - docker.io
          - docker-compose
        state: present
        update_cache: yes
    - name: Clone SocialPredict repo
      git:
        repo: "https://github.com/openpredictionmarkets/socialpredict.git"
        dest: "/opt/socialpredict"
        version: main
        force: yes
Verify Access

Run:

bash
Copy
Edit
ansible-playbook -i deploy/ansible/hosts.ini deploy/ansible/setup.yml
Confirm repo exists at /opt/socialpredict.

Project Plan 2: Add Maintenance Tasks via Ansible
Goal
Automate log rotation, container pruning, and other cleanup tasks to prevent disk space issues on the Droplet.

Steps
Update Docker to Use Log Rotation

Extend the playbook to create /etc/docker/daemon.json:

yaml
Copy
Edit
- name: Configure Docker log rotation
  copy:
    dest: /etc/docker/daemon.json
    content: |
      {
        "log-driver": "json-file",
        "log-opts": {
          "max-size": "10m",
          "max-file": "3"
        }
      }
  notify: Restart Docker
Add handler:

yaml
Copy
Edit
handlers:
  - name: Restart Docker
    service:
      name: docker
      state: restarted
Set Up Cron Jobs

Weekly prune (Sunday 3 AM):

yaml
Copy
Edit
- name: Weekly Docker prune
  cron:
    name: "Docker System Prune"
    minute: "0"
    hour: "3"
    weekday: "0"
    job: "docker system prune -af --volumes >/dev/null 2>&1"
Daily log truncation:

yaml
Copy
Edit
- name: Daily truncate container logs
  cron:
    name: "Truncate Docker Logs"
    minute: "0"
    hour: "2"
    job: "find /var/lib/docker/containers/ -name '*-json.log' -exec truncate -s 0 {} \\;"
Test Maintenance

Deploy playbook:

bash
Copy
Edit
ansible-playbook -i deploy/ansible/hosts.ini deploy/ansible/setup.yml
Verify:

bash
Copy
Edit
crontab -l
docker info | grep 'Log Driver'
Project Plan 3: CI/CD with GitHub Pipelines (Future)
Goal
Deploy SocialPredict automatically to DigitalOcean whenever a branch is pushed or merged (e.g., to main or staging).

Steps
Secrets Setup

Add these as GitHub Secrets:

DOCKER_REGISTRY (if using GHCR)

DOCKER_USERNAME

DOCKER_PASSWORD (or PAT for GHCR)

DO_SSH_KEY (private key for 143.198.177.112)

Workflow Skeleton (.github/workflows/deploy.yml)

yaml
Copy
Edit
name: Deploy to Droplet

on:
  push:
    branches: [ "main" ]

jobs:
  build-deploy:
    runs-on: ubuntu-latest
    steps:
      - name: Checkout
        uses: actions/checkout@v4

      - name: Build Docker images
        run: |
          docker build -t socialpredict-backend ./backend
          docker build -t socialpredict-frontend ./frontend

      - name: Push images (optional)
        if: env.DOCKER_REGISTRY != ''
        run: |
          echo "${{ secrets.DOCKER_PASSWORD }}" | docker login ${{ secrets.DOCKER_REGISTRY }} -u ${{ secrets.DOCKER_USERNAME }} --password-stdin
          docker push ${{ secrets.DOCKER_REGISTRY }}/socialpredict-backend:latest
          docker push ${{ secrets.DOCKER_REGISTRY }}/socialpredict-frontend:latest

      - name: Deploy via SSH
        uses: appleboy/ssh-action@master
        with:
          host: 143.198.177.112
          username: root
          key: ${{ secrets.DO_SSH_KEY }}
          script: |
            cd /opt/socialpredict
            git pull origin main
            ./SocialPredict up
Decide: Local vs Registry Builds

Initially: build on Droplet (simpler, no registry).

Later: Switch to build on CI, push to GHCR, and docker pull on Droplet.

