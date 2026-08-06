# Notifier Frontend UI/UX Refactoring Plan

## Phase 0: Audit Results

### Routes (22 page.tsx files)
| Route | Status | Issues |
|-------|--------|--------|
| `app/page.tsx` | Redirect | OK |
| `[locale]/page.tsx` | Redirect | OK |
| `[locale]/dashboard/` | Active | Uses local metric-card |
| `[locale]/deliveries/` | Active | Needs table standardization |
| `[locale]/deliveries/[id]/` | Active | Detail page |
| `[locale]/notifications/` | Active | List page |
| `[locale]/notifications/new/` | Active | Form page |
| `[locale]/notifications/[id]/` | Active | Detail page |
| `[locale]/observability/` | Active | Metrics dashboard |
| `[locale]/preferences/` | Active | Settings page |
| `[locale]/providers/` | Active | List page |
| `[locale]/providers/new/` | Active | Form page |
| `[locale]/providers/[providerId]/` | Active | Detail page |
| `[locale]/providers/[providerId]/edit/` | Active | Edit form |
| `[locale]/reminders/` | Active | List page |
| `[locale]/reminders/new/` | Active | Form page |
| `[locale]/reminders/[id]/` | Active | Detail page |
| `[locale]/settings/` | Active | Settings page |
| `[locale]/templates/` | Active | List page |
| `[locale]/templates/new/` | Active | Form page |
| `[locale]/templates/[id]/` | Active | Detail page |
| `[locale]/tenants/` | Active | List page |

### Duplicate UI Components (23 files to remove)
All 23 files under `src/components/ui/` are duplicates of `@minisource/ui`:
alert-dialog, alert, avatar, badge, button, card, command, dialog, dropdown-menu,
input, label, popover, scroll-area, select, separator, sheet, skeleton, sonner,
switch, table, tabs, textarea, tooltip.

### What's Already Good
- App shell uses `@minisource/app-shell` (SidebarProvider, SidebarInset, Topbar)
- Navigation is typed and grouped
- Auth guard + provider pattern
- i18n with next-intl
- Shared error normalization (just added)

### Key Visual Issues
1. globals.css has redundant CSS variable definitions (duplicating design-system tokens)
2. Some pages use local component imports instead of shared patterns
3. Status badges and empty/error states need standardization

## Phase 1: Foundation
- [x] Fix i18n default locale (fa→en) - DONE
- [ ] Replace duplicate UI components with @minisource/ui imports
- [ ] Clean globals.css (remove redundant CSS variables, keep only notifier-specific)
- [ ] Verify shared patterns (page-container, page-header, empty/error/loading states)

## Phase 2: Application Shell
- App shell already uses @minisource/app-shell - NO MAJOR CHANGES NEEDED
- Verify sidebar active state matching
- Ensure mobile navigation works

## Phase 3: Shared UI Patterns  
- Standardize page-header across all routes
- Standardize forms (use @minisource/rhf where possible)
- Standardize tables/lists
- Ensure all routes have proper loading/empty/error states

## Phase 4: Route-by-Route
- Audit each page for consistency
- Fix imports to use @minisource/ui

## Phase 5: Cleanup
- Remove duplicate UI components after all consumers migrated
- Remove dead code

## Phase 6: Validation
- Typecheck
- Lint
- Build
