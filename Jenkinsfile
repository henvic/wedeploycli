import net.sf.json.JSONArray;
import net.sf.json.JSONObject;

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
          sh './.jenkins/test.sh'
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

void buildStep(String message, Closure closure) {
  try {
    setBuildStatus(message, "PENDING");

    closure();

    setBuildStatus(message, "SUCCESS");
  }
  catch (Exception e) {
    setBuildStatus(message, "FAILURE");
    throw e
  }
}

String getGitAuthor() {
  def commit = sh(returnStdout: true, script: 'git rev-parse HEAD')
  return sh(returnStdout: true, script: "git --no-pager show -s --format='%an' ${commit}").trim()
}

String getLastCommitMessage() {
  return sh(returnStdout: true, script: 'git log -1 --pretty=%B').trim()
}

String getRandom(String[] array) {
    int rnd = new Random().nextInt(array.length);
    return array[rnd];
}

void pushFailureToSlack(step) {
  String[] errorMessages = [
    'Hey, Vader seems to be mad at you.',
    'Please! Don\'t break the CI ;/',
    'Houston, we have a problem'
  ];

  String title = "FAILED: Job ${env.JOB_NAME} - ${env.BUILD_NUMBER}";
  String text = getRandom(errorMessages);

  JSONArray attachments = buildAttachments(
    "BUILD FAILED: ${step} - wedeploy/cli-functional-tests",
    text,
    'CI BUILD FAILED',
    title,
    '#ff0000'
  );

  slackSend (color: '#ff0000', attachments: attachments.toString());
}

void pushSuccessToSlack() {
  String[] successMessages = [
    'Howdy, we\'re back on track.',
    'YAY!',
    'The force is strong with this one.'
  ];

  String title = "BUILD FIXED: Job ${env.JOB_NAME} - ${env.BUILD_NUMBER}";
  String text = getRandom(successMessages);

  JSONArray attachments = buildAttachments(
    'BUILD FIXED - wedeploy/api',
    text,
    'CI BUILD FIXED',
    title,
    '#5fba7d'
  );

  slackSend (color: '#5fba7d', attachments: attachments.toString());
}

void setBuildStatus(String message, String state) {
  step([
      $class: "GitHubCommitStatusSetter",
      reposSource: [$class: "ManuallyEnteredRepositorySource", url: "https://github.com/wedeploy/cli-functional-tests"],
      contextSource: [$class: "ManuallyEnteredCommitContextSource", context: "ci/jenkins/build-status"],
      errorHandlers: [[$class: "ChangingBuildStatusErrorHandler", result: "UNSTABLE"]],
      statusResultSource: [ $class: "ConditionalStatusResultSource", results: [[$class: "AnyBuildResult", message: message, state: state]] ]
  ]);
}
