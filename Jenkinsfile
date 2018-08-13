@Library('ReleaseManagementSharedLib') _

pipeline {
  agent {
    node {
      label 'cli'
    }
  }

  triggers { cron("H 12 * * 1-7") }

  stages {
    stage('Setup') {
      steps {
        buildStep('Setup') {
          sh './.jenkins/main.sh --setup-machine'
          sh './.jenkins/main.sh --shutdown-infrastructure'
        }
      }
    }
    stage('Pull Infrastructure Images') {
      steps {
        buildStep('Pull Infrastructure Images') {
          sh './.jenkins/main.sh --pull-infrastructure-images'
        }
      }
    }
    stage('Start Infrastructure') {
      steps {
        buildStep('Start Infrastructure') {
          sh './.jenkins/main.sh --start-infrastructure'
        }
      }
    }
    stage('Create Test User') {
      steps {
        buildStep('Create Test User') {
          sh './.jenkins/main.sh --create-test-user'
        }
      }
    }
    stage('Tests') {
      steps {
        buildStep('Tests') {
          timeout(time: 10, unit: 'MINUTES') {
            sh './.jenkins/test.sh'
          }
        }
      }
    }
  }
  post {
    always {
      junit(allowEmptyResults: true, testResults: 'functional/test-results/TEST-*.xml')

      archiveArtifacts artifacts: 'functional/test-results/report.txt'

      sh './.jenkins/main.sh --shutdown-infrastructure'
    }
  }
}