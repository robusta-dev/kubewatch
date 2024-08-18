docker buildx build \
  --build-arg BUILDKIT_INLINE_CACHE=1 \
  --platform linux/arm64,linux/amd64 \
  --tag robustadev/kubewatch:${TAG} \
  --tag us-central1-docker.pkg.dev/genuine-flight-317411/devel/kubewatch:${TAG} \
  --push \
  .