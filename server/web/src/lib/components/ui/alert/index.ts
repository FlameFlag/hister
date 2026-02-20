import Root from './alert.svelte';
import Description from './alert-description.svelte';
import Title from './alert-title.svelte';
import { alertVariants, type AlertVariant } from './alert.svelte';
export { alertVariants, type AlertVariant };

export {
  Root,
  Description,
  Title,
  //
  Root as Alert,
  Description as AlertDescription,
  Title as AlertTitle,
};
