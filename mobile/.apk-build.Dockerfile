FROM ubuntu:24.04

ENV DEBIAN_FRONTEND=noninteractive

RUN apt-get update \
 && apt-get install -y --no-install-recommends \
      openjdk-21-jdk-headless unzip curl ca-certificates \
 && rm -rf /var/lib/apt/lists/*

ENV ANDROID_HOME=/opt/android-sdk
ENV ANDROID_SDK_ROOT=/opt/android-sdk
ENV GRADLE_HOME=/opt/gradle
ENV PATH=$PATH:/opt/gradle/bin:/opt/android-sdk/cmdline-tools/latest/bin:/opt/android-sdk/platform-tools

RUN mkdir -p $ANDROID_HOME/cmdline-tools \
 && curl -fsSL --retry 8 --retry-delay 5 --retry-all-errors -C - \
      -o /tmp/tools.zip https://dl.google.com/dl/android/repository/commandlinetools-linux-13114758_latest.zip \
 && unzip -q /tmp/tools.zip -d $ANDROID_HOME/cmdline-tools \
 && mv $ANDROID_HOME/cmdline-tools/cmdline-tools $ANDROID_HOME/cmdline-tools/latest \
 && rm /tmp/tools.zip \
 && (yes | sdkmanager --licenses > /dev/null 2>&1 || true) \
 && sdkmanager "platform-tools" "platforms;android-35" "build-tools;35.0.0"

RUN curl -fsSL --retry 8 --retry-delay 5 --retry-all-errors -C - \
      -o /tmp/gradle.zip https://services.gradle.org/distributions/gradle-8.11.1-bin.zip \
 && unzip -q /tmp/gradle.zip -d /opt \
 && mv /opt/gradle-8.11.1 /opt/gradle \
 && rm /tmp/gradle.zip

# Pas de Node ici : le build web et `cap sync` tournent sur l'hôte (npx cap sync),
# le conteneur ne sert qu'à Gradle, comme pour ticket-park.
WORKDIR /app
