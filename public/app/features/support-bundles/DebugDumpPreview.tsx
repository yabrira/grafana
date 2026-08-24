import { css } from '@emotion/css';

import { type GrafanaTheme2 } from '@grafana/data';
import { useStyles2 } from '@grafana/ui';

interface DebugDumpPreviewProps {
  html: string;
  title?: string;
}

export function DebugDumpPreview({ html, title = 'Support dump' }: DebugDumpPreviewProps) {
  const styles = useStyles2(getStyles);

  return (
    <section className={styles.wrap}>
      <h2>{title}</h2>
      <div className={styles.body} dangerouslySetInnerHTML={{ __html: html }} />
    </section>
  );
}

const getStyles = (theme: GrafanaTheme2) => ({
  wrap: css({
    padding: theme.spacing(2),
  }),
  body: css({
    fontFamily: theme.typography.fontFamilyMonospace,
    whiteSpace: 'pre-wrap',
  }),
});
