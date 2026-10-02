allprojects {
    repositories {
        google()
        mavenCentral()
        // Espressif publishes the Android provisioning SDK on JitPack.
        // Restrict the repository to that group so other dependencies keep
        // resolving from Google Maven and Maven Central.
        maven {
            url = uri("https://jitpack.io")
            content {
                includeGroup("com.github.espressif")
            }
        }
    }
}

val newBuildDir: Directory =
    rootProject.layout.buildDirectory
        .dir("../../build")
        .get()
rootProject.layout.buildDirectory.value(newBuildDir)

subprojects {
    val newSubprojectBuildDir: Directory = newBuildDir.dir(project.name)
    project.layout.buildDirectory.value(newSubprojectBuildDir)
}
subprojects {
    project.evaluationDependsOn(":app")
}

tasks.register<Delete>("clean") {
    delete(rootProject.layout.buildDirectory)
}
