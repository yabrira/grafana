import { css } from '@emotion/css';

import { type GrafanaTheme2 } from '@grafana/data';
import { t } from '@grafana/i18n';
import { Alert, useStyles2 } from '@grafana/ui';

/** Hardcoded walkthrough numbers — not loaded from billing or quota APIs. */
export const SOFT_SPEND_LIMIT_FIXTURE = {
  usedPercent: 82,
  usedAmount: 'SAR 41,000',
  limitAmount: 'SAR 50,000',
} as const;

/**
 * Awareness-only soft spend / quota warning for the ACDY-18 login demo.
 * Does not block login or payments; approvers get a heads-up only.
 */
export function SoftSpendLimitBanner() {
  const styles = useStyles2(getStyles);

  return (
    <Alert
      className={styles.banner}
      severity="warning"
      title={t('login.soft-spend-limit.title', 'Soft spend limit')}
    >
      {t(
        'login.soft-spend-limit.body',
        "You're at {{usedPercent}}% of this month's soft limit ({{usedAmount}} / {{limitAmount}}). Payments still go through. Approvers get a heads-up — not a hard block.",
        SOFT_SPEND_LIMIT_FIXTURE
      )}
    </Alert>
  );
}

const getStyles = (theme: GrafanaTheme2) => {
  return {
    banner: css({
      width: '100%',
      marginBottom: theme.spacing(2),
      textAlign: 'left',
    }),
  };
};
