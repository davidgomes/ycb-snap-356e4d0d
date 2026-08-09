#pragma once
#include "envoy/common/time.h"

#include "source/common/common/lock_guard.h"
#include "source/extensions/common/aws/credentials_provider.h"

namespace Envoy {
namespace Extensions {
namespace Common {
namespace Aws {

class CachedCredentialsProviderBase : public CredentialsProvider,
                                      public Logger::Loggable<Logger::Id::aws> {
public:
  Credentials getCredentials() override {
    Thread::LockGuard guard(mu_);
    refreshIfNeeded();
    return cached_credentials_;
  }
  bool credentialsPending() override { return false; };

protected:
  Thread::MutexBasicLockable mu_;
  SystemTime last_updated_ ABSL_GUARDED_BY(mu_);
  Credentials cached_credentials_ ABSL_GUARDED_BY(mu_);

  void refreshIfNeeded() ABSL_EXCLUSIVE_LOCKS_REQUIRED(mu_) {
    if (needsRefresh()) {
      refresh();
    }
  }

  virtual bool needsRefresh() ABSL_EXCLUSIVE_LOCKS_REQUIRED(mu_) PURE;
  virtual void refresh() ABSL_EXCLUSIVE_LOCKS_REQUIRED(mu_) PURE;
};

} // namespace Aws
} // namespace Common
} // namespace Extensions
} // namespace Envoy
