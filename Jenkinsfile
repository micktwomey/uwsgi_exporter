node {
    stage('Checkout'){
      deleteDir()
      checkout scm
    } 

    def pwd=pwd()

    stage ('Package') {
       '''#!/bin/bash
       make build-deb
       cp uwsgi-exporter_0.1_amd64.deb uwsgi-export_0.1_amd64-xenial.deb
        '''
    }

  // Publish as a Ubuntu Xenial package 
  stage('Publish') {
    archive '*.deb'
    build(
      job: 'Publish/publish-ubuntu-package', 
      parameters: [
        string(name: 'PROJECT_NAME', value: env.JOB_NAME), 
        string(name: 'CODENAME', value: 'xenial')
      ], 
      wait: false,
    )
  }
}
