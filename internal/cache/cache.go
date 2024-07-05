package cache

import "github.com/zhex/local-bbp/internal/models"

var defaultCaches = models.Caches{}

func init() {
	defaultCaches.Set("node", models.NewCache("node_modules", []string{"package.json"}))
	defaultCaches.Set("gradle", models.NewCache("～/.gradle/cache", []string{"build.gradle.kts", "settings.gradle.kts"}))
	defaultCaches.Set("maven", models.NewCache("~/.m2/repository", []string{"pom.xml"}))
	defaultCaches.Set("pip", models.NewCache("~/.cache/pip", []string{"requirements.txt"}))
	defaultCaches.Set("composer", models.NewCache("~/.composer/cache", []string{"composer.json"}))
	defaultCaches.Set("dotnetcore", models.NewCache("~/.nuget/packages", []string{"packages.config"}))
	defaultCaches.Set("composer", models.NewCache("~/.composer/cache", []string{"composer.json"}))
	defaultCaches.Set("ivy", models.NewCache("~/.ivy2/cache", []string{"ivy.xml"}))
	defaultCaches.Set("sbt", models.NewCache("~/.sbt", []string{"build.sbt"}))
}
