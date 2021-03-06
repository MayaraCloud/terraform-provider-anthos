# This makefile is meant to build the Anthos Terraform provider
# 2020 Dario Ferrer <dario@mayara.io>
#
# Requirements:
#   * GNU make


all: plan

plan: init
	GO_DEBUG=true terraform plan -out /tmp/tf.plan

destroy: init
	GO_DEBUG=true terraform plan -destroy -out /tmp/tf.plan

debug: init
	TF_LOG=DEBUG terraform plan -out /tmp/tf.plan

init: build
	terraform init

build:
	go build -o terraform-provider-anthos

apply:
	GO_DEBUG=true terraform apply /tmp/tf.plan
