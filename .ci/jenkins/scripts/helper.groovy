openshift = null
container = null
properties = null

defaultImageParamsPrefix = 'IMAGE'
baseImageParamsPrefix = 'BASE_IMAGE'
promoteImageParamsPrefix = 'PROMOTE_IMAGE'

void initPipeline() {
    properties = load '.ci/jenkins/scripts/properties.groovy'

    openshift = load '.ci/jenkins/scripts/openshift.groovy'
    openshift.openshiftApiKey = env.OPENSHIFT_API_KEY
    openshift.openshiftApiCredsKey = env.OPENSHIFT_CREDS_KEY

    container = load '.ci/jenkins/scripts/container.groovy'
    container.containerEngine = env.CONTAINER_ENGINE
    container.containerTlsOptions = env.CONTAINER_TLS_OPTIONS
    container.containerOpenshift = openshift
}

void updateDisplayName() {
    if (params.DISPLAY_NAME) {
        currentBuild.displayName = params.DISPLAY_NAME
    }
}

void cleanGoPath() {
    sh 'rm -rf $GOPATH/bin/*'
}

String buildTempOpenshiftImageFullName(boolean internal=false) {
    return "${getTempOpenshiftImageName(internal)}:${getTempTag()}"
}

String getTempOpenshiftImageName(boolean internal=false) {
    String registry = internal ? openshiftInternalRegistry : openshift.getOpenshiftRegistry()
    return "${registry}/openshift/${env.OPERATOR_IMAGE_NAME}"
}

String getTempTag() {
    return "pr-${getShortGitCommitHash()}"
}

void checkoutRepo(String repoName = '', String directory = '') {
    repoName = repoName ?: getRepoName()
    closure = {
        deleteDir()
        checkout(githubscm.resolveRepository(repoName, getGitAuthor(), getBuildBranch(), false))
        // need to manually checkout branch since on a detached branch after checkout command
        sh "git checkout ${getBuildBranch()}"
    }

    if (directory) {
        dir(directory, closure)
    } else {
        closure()
    }
}

void loginRegistry(String paramsPrefix = defaultImageParamsPrefix) {
    if (isImageInOpenshiftRegistry(paramsPrefix)) {
        container.loginOpenshiftRegistry()
    } else if (getImageRegistryCredentials(paramsPrefix)) {
        container.loginContainerRegistry(getImageRegistry(paramsPrefix), getImageRegistryCredentials(paramsPrefix))
    }
}

void installGitHubReleaseCLI() {
    sh 'go install github.com/github-release/github-release@latest'
}

void createRelease() {
    if (isReleaseExist()) {
        deleteRelease()
    }

    if (githubscm.isTagExist('origin', getGitTag())) {
        githubscm.removeLocalTag(getGitTag())
        githubscm.removeRemoteTag('origin', getGitTag(), getGitAuthorCredsID())
    }

    sh "make build-cli release=true version=${getProjectVersion()}"
    def releaseName = "Kogito Operator and CLI Version ${getProjectVersion()}"
    def description = 'Kogito Operator is a Kubernetes based operator for Kogito Runtimes\' deployment from the source. Additionally, to facilitate interactions with the operator, we also offer a CLI (Command Line Interface) to deploy Kogito applications.'
    def releasePath = 'build/_output/release/'
    def cliBaseName = "kogito-cli-${getProjectVersion()}"
    def darwinFileName = "${cliBaseName}-darwin-amd64.tar.gz"
    def linuxFileName = "${cliBaseName}-linux-amd64.tar.gz"
    def windowsFileName = "${cliBaseName}-windows-amd64.zip"
    def yamlInstaller = 'kogito-operator.yaml'
    withCredentials([string(credentialsId: env.GITHUB_TOKEN_CREDS_ID, variable: 'GITHUB_TOKEN')]) {
        sh """
            export GITHUB_USER=${getGitAuthor()}
            github-release release --tag ${getGitTag()} --target \"${getBuildBranch()}\" --name \"${releaseName}\" --description \"${description}\" --pre-release
            github-release upload --tag ${getGitTag()} --name \"${darwinFileName}\" --file \"${releasePath}${darwinFileName}\"
            github-release upload --tag ${getGitTag()} --name \"${linuxFileName}\" --file \"${releasePath}${linuxFileName}\"
            github-release upload --tag ${getGitTag()} --name \"${windowsFileName}\" --file \"${releasePath}${windowsFileName}\"
            github-release upload --tag ${getGitTag()} --name \"${yamlInstaller}\" --file \"${yamlInstaller}\"
        """
    }
}

boolean isReleaseExist() {
    releaseExistStatus = -1
    withCredentials([string(credentialsId: env.GITHUB_TOKEN_CREDS_ID, variable: 'GITHUB_TOKEN')]) {
        releaseExistStatus = sh(returnStatus: true, script: """
            export GITHUB_USER=${getGitAuthor()}
            github-release info --tag ${getGitTag()}
        """)
    }
    return releaseExistStatus == 0
}

void deleteRelease() {
    withCredentials([string(credentialsId: env.GITHUB_TOKEN_CREDS_ID, variable: 'GITHUB_TOKEN')]) {
        sh """
            export GITHUB_USER=${getGitAuthor()}
            github-release delete --tag ${getGitTag()}
        """
    }
}

// Set images public on quay. Useful when new images are introduced.
void makeQuayImagePublic(String repository, String paramsPrefix = defaultImageParamsPrefix) {
    String namespace = getImageNamespace(paramsPrefix)
    echo "Check and set public if needed Quay repository ${namespace}/${repository}"
    try {
        cloud.makeQuayImagePublic(namespace, repository, [ usernamePassword: getImageRegistryCredentials(paramsPrefix)])
    } catch (err) {
        echo "[ERROR] Cannot set image quay.io/${namespace}/${repository} as visible"
    }
}

String getPropertiesImagePrefix() {
    return 'images'
}

String getImageRegistryProperty() {
    return contructImageProperty('registry')
}

String getImageNamespaceProperty() {
    return contructImageProperty('namespace')
}

String getImageNamePrefixProperty() {
    return contructImageProperty('name-prefix')
}

String getImageNameSuffixProperty() {
    return contructImageProperty('name-suffix')
}

String getImageNamesProperty() {
    return contructImageProperty('names')
}

String getImageTagProperty() {
    return contructImageProperty('tag')
}

String contructImageProperty(String suffix) {
    return "${getPropertiesImagePrefix()}.${suffix}"
}

////////////////////////////////////////////////////////////////////////
// Image information
////////////////////////////////////////////////////////////////////////

boolean isImageInOpenshiftRegistry(String paramsPrefix = defaultImageParamsPrefix) {
    return params[constructKey(paramsPrefix, 'USE_OPENSHIFT_REGISTRY')]
}

String getImageRegistryCredentials(String paramsPrefix = defaultImageParamsPrefix) {
    return isImageInOpenshiftRegistry(paramsPrefix) ? '' : params[constructKey(paramsPrefix, 'REGISTRY_CREDENTIALS')]
}

String getImageRegistry(String paramsPrefix = defaultImageParamsPrefix) {
    if (isImageInOpenshiftRegistry(paramsPrefix)) {
        return openshift.getOpenshiftRegistry()
    } else if (paramsPrefix == baseImageParamsPrefix && properties.contains(getImageRegistryProperty())) {
        return properties.retrieve(getImageRegistryProperty())
    }
    return  params[constructKey(paramsPrefix, 'REGISTRY')]
}

String getImageNamespace(String paramsPrefix = defaultImageParamsPrefix) {
    if (isImageInOpenshiftRegistry(paramsPrefix)) {
        return 'openshift'
    } else if (paramsPrefix == baseImageParamsPrefix && properties.contains(getImageNamespaceProperty())) {
        return properties.retrieve(getImageNamespaceProperty())
    }
    return params[constructKey(paramsPrefix, 'NAMESPACE')]
}

String getImageNamePrefix(String paramsPrefix = defaultImageParamsPrefix) {
    if (paramsPrefix == baseImageParamsPrefix && properties.contains(getImageNamePrefixProperty())) {
        return properties.retrieve(getImageNamePrefixProperty())
    }
    return params[constructKey(paramsPrefix, 'NAME_PREFIX')]
}

List getImageNames(String paramsPrefix = defaultImageParamsPrefix) {
    String commaSepImages = ''
    if (paramsPrefix == baseImageParamsPrefix && properties.contains(getImageNamesProperty())) {
        commaSepImages = properties.retrieve(getImageNamesProperty())
    } else {
        commaSepImages = params[constructKey(paramsPrefix, 'NAMES')]
    }
    return commaSepImages.split(',') as List
}

String getImageNameSuffix(String paramsPrefix = defaultImageParamsPrefix) {
    if (paramsPrefix == baseImageParamsPrefix && properties.contains(getImageNameSuffixProperty())) {
        return properties.retrieve(getImageNameSuffixProperty())
    }
    return params[constructKey(paramsPrefix, 'NAME_SUFFIX')]
}

String getFullImageName(String imageName, String paramsPrefix = defaultImageParamsPrefix) {
    prefix = getImageNamePrefix(paramsPrefix)
    suffix = getImageNameSuffix(paramsPrefix)
    return (prefix ? prefix + '-' : '') + imageName + (suffix ? '-' + suffix : '')
}

String getImageTag(String paramsPrefix = defaultImageParamsPrefix) {
    if (paramsPrefix == baseImageParamsPrefix && properties.contains(getImageTagProperty())) {
        return properties.retrieve(getImageTagProperty())
    }
    return params[constructKey(paramsPrefix, 'TAG')] ?: getShortGitCommitHash()
}

String getImageFullTag(String imageName, String paramsPrefix = defaultImageParamsPrefix, String tag = '') {
    String fullTag = getImageRegistry(paramsPrefix)
    fullTag += "/${getImageNamespace(paramsPrefix)}"
    fullTag += "/${getFullImageName(imageName, paramsPrefix)}"
    fullTag += ":${tag ?: getImageTag(paramsPrefix)}"
    return fullTag
}

String constructKey(String keyPrefix, String key) {
    return keyPrefix ? "${keyPrefix}_${key}" : key
}

String getShortGitCommitHash() {
    return sh(returnStdout: true, script: 'git rev-parse --short HEAD').trim()
}

String getReducedTag(String paramsPrefix = defaultImageParamsPrefix) {
    String tag = helper.getImageTag(paramsPrefix)
    try {
        String[] versionSplit = tag.split('\\.')
        return "${versionSplit[0]}.${versionSplit[1]}"
    } catch (error) {
        echo "${tag} cannot be reduced to the format X.Y"
    }
    return ''
}

/////////////////////////////////////////////////////////////////////
// Utils

boolean isRelease() {
    return env.RELEASE && env.RELEASE.toBoolean()
}

boolean isCreatePr() {
    return params.CREATE_PR
}

String getRepoName() {
    return env.REPO_NAME
}

String getBuildBranch() {
    return params.BUILD_BRANCH_NAME
}

String getGitAuthor() {
    return "${GIT_AUTHOR}"
}

String getGitAuthorCredsID() {
    return env.AUTHOR_CREDS_ID
}

String getBotBranch() {
    return "${getProjectVersion()}-${env.BOT_BRANCH_HASH}"
}

String getBotAuthor() {
    return env.GIT_AUTHOR_BOT
}

String getBotAuthorCredsID() {
    return env.BOT_CREDENTIALS_ID
}

String getProjectVersion() {
    return properties.retrieve('project.version') ?: params.PROJECT_VERSION
}

String getNextVersion() {
    return util.getNextVersion(getProjectVersion(), 'micro', 'snapshot')
}

String getSnapshotBranch() {
    return "${getNextVersion()}-${env.BOT_BRANCH_HASH}"
}

boolean shouldLaunchTests() {
    return !params.SKIP_TESTS
}

String getDeployPropertiesFileUrl() {
    String url = params.DEPLOY_BUILD_URL
    if (url) {
        return "${url}${url.endsWith('/') ? '' : '/'}artifact/${env.PROPERTIES_FILE_NAME}"
    }
    return ''
}

String getGitTag() {
    return params.GIT_TAG != '' ? params.GIT_TAG : "v${getProjectVersion()}"
}

boolean isDeployLatestTag() {
    return params.DEPLOY_WITH_LATEST_TAG
}

/////////////////////////////////////////////////////////////////////
// BDD

Map getBDDCommonParameters(boolean runtime_app_registry_internal) {
    Map testParamsMap = [:]

    testParamsMap['load_default_config'] = true
    testParamsMap['ci'] = 'jenkins'

    testParamsMap['operator_image_tag'] = "${getTempOpenshiftImageName(true)}:${getTempTag()}"

    String mavenRepository = env.MAVEN_ARTIFACT_REPOSITORY ?: (isRelease() ? env.DEFAULT_STAGING_REPOSITORY : '')
    if (mavenRepository) {
        // No mirror if we set directly the Maven repository
        // Tests will be slower but we need to test against specific artifacts
        testParamsMap['custom_maven_repo_url'] = mavenRepository
        testParamsMap['maven_ignore_self_signed_certificate'] = true
    }
    // Disabled as we now use IBMCloud
    // Follow-up issue to make it more dynamic: https://issues.redhat.com/browse/KOGITO-5739
    // if (env.MAVEN_MIRROR_REPOSITORY) {
    //     testParamsMap['maven_mirror_url'] = env.MAVEN_MIRROR_REPOSITORY
    //     testParamsMap['maven_ignore_self_signed_certificate'] = true
    // }

    if (params.EXAMPLES_REF) {
        testParamsMap['examples_ref'] = params.EXAMPLES_REF
    }
    if (params.EXAMPLES_URI) {
        testParamsMap['examples_uri'] = params.EXAMPLES_URI
    }

    if (params.NATIVE_BUILDER_IMAGE) {
        testParamsMap['native_builder_image'] = params.NATIVE_BUILDER_IMAGE
    }

    // Clean the cluster before/after BDD test execution
    testParamsMap['enable_clean_cluster'] = true

    testParamsMap['container_engine'] = containerEngine

    return testParamsMap
}

Map getBDDBuildImageParameters(String paramsPrefix = defaultImageParamsPrefix) {
    Map testParamsMap = [:]

    String registry = "${getImageRegistry(paramsPrefix)}/${getImageNamespace(paramsPrefix)}"
    String nameSuffix = getImageNameSuffix(paramsPrefix) ? "-${getImageNameSuffix(paramsPrefix)}" : ''
    String tag = getImageTag(paramsPrefix) ? ":${getImageTag(paramsPrefix)}" : ''

    testParamsMap['build_builder_image_tag'] = "${registry}/kogito-builder${nameSuffix}${tag}"
    testParamsMap['build_runtime_jvm_image_tag'] = "${registry}/kogito-runtime-jvm${nameSuffix}${tag}"
    testParamsMap['build_runtime_native_image_tag'] = "${registry}/kogito-runtime-native${nameSuffix}${tag}"

    return testParamsMap
}

Map getBDDServicesImageParameters(String paramsPrefix = defaultImageParamsPrefix) {
    Map testParamsMap = [:]

    testParamsMap['services_image_registry'] = "${getImageRegistry(paramsPrefix)}/${getImageNamespace(paramsPrefix)}"
    testParamsMap['services_image_name_suffix'] = getImageNameSuffix(paramsPrefix) ?: ''
    testParamsMap['services_image_version'] = getImageTag(paramsPrefix) ?: ''

    return testParamsMap
}

Map getBDDRuntimeImageParameters(String paramsPrefix = defaultImageParamsPrefix) {
    Map testParamsMap = [:]

    testParamsMap['runtime_application_image_registry'] = "${getImageRegistry(paramsPrefix)}/${getImageNamespace(paramsPrefix)}"
    testParamsMap['runtime_application_image_name_prefix'] = getImageNamePrefix(paramsPrefix) ?: ''
    testParamsMap['runtime_application_image_name_suffix'] = getImageNameSuffix(paramsPrefix) ?: ''
    testParamsMap['runtime_application_image_version'] = getImageTag(paramsPrefix) ?: ''

    return testParamsMap
}

String getNativeTag() {
    return '@native'
}

String getNonNativeTag() {
    return "~${getNativeTag()}"
}

return this
