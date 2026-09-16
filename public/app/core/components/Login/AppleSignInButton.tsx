import { useState } from 'react';

import { Trans, t } from '@grafana/i18n';
import { Alert, Button, Icon, useStyles2, useTheme2 } from '@grafana/ui';

import { getButtonStyleFor, getServiceStyles } from './LoginServiceButtons';

const APPLE_BG_COLOR = '#000000';

/**
 * Placeholder affordance for Apple sign-in. There is no Apple OAuth connector on the backend, so
 * this renders a Button rather than a LinkButton: it must not navigate to `login/apple`, which
 * would 404 and surface as a generic login failure.
 */
export const AppleSignInButton = () => {
  const [showNotConfigured, setShowNotConfigured] = useState(false);
  const theme = useTheme2();
  const styles = useStyles2(getServiceStyles);

  return (
    <>
      <Button
        className={getButtonStyleFor(APPLE_BG_COLOR, styles, theme)}
        onClick={() => setShowNotConfigured(true)}
        fullWidth
      >
        <Icon className={styles.buttonIcon} name="signin" />
        <Trans i18nKey="login.apple.sign-in">Sign in with Apple</Trans>
      </Button>
      {showNotConfigured && (
        <Alert severity="info" title={t('login.apple.not-configured-title', 'Sign in with Apple is not available')}>
          <Trans i18nKey="login.apple.not-configured-body">
            This sign-in method is not configured on this Grafana instance. Use one of the other sign-in options.
          </Trans>
        </Alert>
      )}
    </>
  );
};
