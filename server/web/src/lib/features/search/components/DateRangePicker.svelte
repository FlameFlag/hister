<script lang="ts">
  import { Calendar as CalendarIcon } from 'lucide-svelte';
  import * as Popover from '$lib/components/ui/popover';
  import { Calendar } from '$lib/components/ui/calendar';
  import { Button } from '$lib/components/ui/button';
  import type { DateValue } from '@internationalized/date';
  import { getLocalTimeZone } from '@internationalized/date';
  import { buttonVariants } from '$lib/components/ui/button';
  import { cn } from '$lib/utils';

  let {
    dateFrom,
    onDateFromChange,
    dateTo,
    onDateToChange,
    minDate,
    maxDate,
  }: {
    dateFrom?: DateValue;
    onDateFromChange?: (value: DateValue | undefined) => void;
    dateTo?: DateValue;
    onDateToChange?: (value: DateValue | undefined) => void;
    minDate: DateValue | undefined;
    maxDate: DateValue | undefined;
  } = $props();

  let popoverOpen = $state(false);
  let calendarValue = $state<DateValue | undefined>();

  function formatCalendarDate(date: DateValue | undefined): string {
    if (!date) return '';
    return date.toDate(getLocalTimeZone()).toLocaleDateString();
  }

  function isDateSelectedFrom(date: DateValue): boolean {
    return dateFrom ? date.compare(dateFrom) === 0 : false;
  }

  function isDateSelectedTo(date: DateValue): boolean {
    return dateTo ? date.compare(dateTo) === 0 : false;
  }

  function isDateDisabled(date: DateValue): boolean {
    if (minDate && date.compare(minDate) < 0) return true;
    if (maxDate && date.compare(maxDate) > 0) return true;
    return false;
  }

  function handleDateClick(date: DateValue) {
    const newValue = date;
    if (!dateFrom) {
      onDateFromChange?.(newValue);
      popoverOpen = false;
    } else if (dateFrom && !dateTo) {
      if (date.compare(dateFrom) >= 0) {
        onDateToChange?.(newValue);
        popoverOpen = false;
      } else {
        onDateFromChange?.(newValue);
        onDateToChange?.(undefined);
      }
    } else {
      onDateFromChange?.(newValue);
      onDateToChange?.(undefined);
    }
    calendarValue = undefined;
  }
</script>

<Popover.Root bind:open={popoverOpen}>
  <Popover.Trigger>
    {#snippet child({ props })}
      <Button
        {...props}
        variant="outline"
        class="
        h-9 cursor-pointer justify-between gap-2
      "
      >
        {#if dateFrom && dateTo}
          {formatCalendarDate(dateFrom)} - {formatCalendarDate(dateTo)}
        {:else if dateFrom}
          {formatCalendarDate(dateFrom)} - Select end date
        {:else}
          Select date range
        {/if}
        <CalendarIcon class="size-4 text-muted-foreground" />
      </Button>
    {/snippet}
  </Popover.Trigger>
  <Popover.Content class="w-auto p-0" align="start">
    <Calendar type="single" bind:value={calendarValue} captionLayout="dropdown">
      {#snippet day({ day: date, outsideMonth })}
        <button
          type="button"
          disabled={isDateDisabled(date)}
          class={cn(
            buttonVariants({ variant: 'ghost' }),
            `
              flex size-(--cell-size) cursor-pointer flex-col items-center
              justify-center gap-1 rounded-md p-0 leading-none font-normal
              whitespace-nowrap select-none
            `,
            (isDateSelectedTo(date) || isDateSelectedFrom(date)) &&
              `
                 bg-primary text-primary-foreground
                 dark:hover:bg-accent/50
               `,
            isDateDisabled(date) && 'cursor-not-allowed opacity-30',
            outsideMonth && 'text-muted-foreground'
          )}
          onclick={() => {
            if (!isDateDisabled(date)) {
              handleDateClick(date);
            }
          }}
        >
          <span class="text-xs opacity-70">{date.day}</span>
        </button>
      {/snippet}
    </Calendar>
  </Popover.Content>
</Popover.Root>
