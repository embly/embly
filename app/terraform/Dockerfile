FROM hashicorp/terraform:0.12.18

WORKDIR /run

ARG AWS_ACCESS_KEY_ID=$AWS_ACCESS_KEY_ID
ARG AWS_SECRET_ACCESS_KEY=$AWS_SECRET_ACCESS_KEY

COPY vpc.tf main.tf ./
RUN terraform init 
COPY entrypoint.sh ./

ENTRYPOINT ["/opt/entrypoint.sh"]