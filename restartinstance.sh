set -e

gcloud compute instances stop remraku --zone us-central1-c
gcloud compute instances start remraku --zone us-central1-c