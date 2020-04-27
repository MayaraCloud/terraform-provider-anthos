# This makefile is meant to build the Anthos Terraform provider
# 2020 Dario Ferrer <dario@mayara.io>
#
# Requirements:
#   * GNU make


all: plan

plan: init
	terraform plan -out /tmp/tf.plan

destroy: init
	terraform plan -destroy -out /tmp/tf.plan

debug: init
	TF_LOG=DEBUG terraform plan -out /tmp/tf.plan

init: build
	terraform init

build:
	go build -o terraform-provider-anthos

apply:
	terraform apply /tmp/tf.plan
