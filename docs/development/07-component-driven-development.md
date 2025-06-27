# Component-Driven Development with Storybook

This document establishes the mandatory workflow for developing UI components using Storybook as the single source of truth for design specifications.

## Philosophy: Storybook as Living Documentation

All UI components and their states must be documented and visually tested through Storybook. This ensures:

- **Design Consistency**: Components match their intended visual specifications
- **Visual Regression Prevention**: Automated visual testing catches unintended changes
- **Component Reusability**: Isolated component development promotes better APIs
- **Design System Enforcement**: All components follow established patterns

## Mandatory Workflow

### 1. Storybook-First Development

**For New Components:**
1. Create the component story BEFORE writing the component implementation
2. Define all important visual states in separate stories
3. Implement the component to match the story specifications
4. Use Storybook's interactive development environment for iteration

**For Component Changes:**
1. Update or add stories to reflect the new states/behavior
2. Verify the visual changes in Storybook before submitting
3. Include story changes in the same PR as component changes

### 2. Story Requirements

Every UI component must have stories that demonstrate:

- **Default state**: The component in its most common configuration
- **Edge cases**: Empty states, loading states, error states
- **Interactive states**: Hover, focus, disabled, active states
- **Data variations**: Different content lengths, data types
- **Responsive behavior**: Different screen sizes (when applicable)

### 3. Story Naming Convention

Stories should follow this naming pattern:
```typescript
// Basic states
export const Default: Story = { ... };
export const Loading: Story = { ... };
export const Error: Story = { ... };

// Data variations
export const LongContent: Story = { ... };
export const EmptyState: Story = { ... };

// Interactive states
export const Disabled: Story = { ... };
export const Active: Story = { ... };

// Context-specific states
export const MobileView: Story = { ... };
export const DarkTheme: Story = { ... };
```

## Visual Regression Testing with Chromatic

### CI Integration

Chromatic runs automatically on every pull request and:
- Captures visual snapshots of all stories
- Compares against the baseline from the main branch
- Flags any visual changes for review
- **Blocks PR merging** if visual changes are detected and not approved

### PR Review Process

1. **Code Review**: Review component implementation and story definitions
2. **Visual Review**: Use Chromatic's preview link to review visual changes
3. **Approval**: Visual changes must be explicitly approved in Chromatic
4. **Merge**: Only after both code and visual approval

## Required Stories by Component Type

### Input Components
- Default state
- Focused state
- Disabled state
- Error state (with validation message)
- Different input lengths

### Interactive Components
- Default state
- Hover state
- Active/pressed state
- Disabled state
- Loading state (if applicable)

### Layout Components
- With minimal content
- With maximum content
- Different breakpoint sizes
- Empty state

### Game-Specific Components
- All possible game states
- Different player counts
- Various data scenarios
- Error/disconnected states

## Pull Request Requirements

### Story Checklist

Every PR involving UI changes must include:

```markdown
### Storybook Checklist

- [ ] **Stories Created/Updated**: All component states have corresponding stories
- [ ] **Visual Review**: Chromatic preview has been reviewed for visual accuracy
- [ ] **Story Completeness**: Stories cover all important states and edge cases
- [ ] **Accessibility**: Stories include accessibility features and testing
- [ ] **Responsive Design**: Stories demonstrate responsive behavior (when applicable)
```

### Chromatic Requirements

- Chromatic CI check must pass
- Any visual changes must be explicitly approved through Chromatic review
- No visual regressions are allowed without justification and approval

## Development Commands

```bash
# Start Storybook development server
npm run storybook

# Build Storybook for production
npm run build-storybook

# Run visual tests locally (requires Chromatic project token)
npm run chromatic
```

## Common Anti-Patterns to Avoid

❌ **Don't**: Implement components without stories
❌ **Don't**: Skip visual testing for "small" changes
❌ **Don't**: Create stories after component implementation
❌ **Don't**: Ignore Chromatic failures or visual differences

✅ **Do**: Start with stories, implement to match
✅ **Do**: Test all component states visually
✅ **Do**: Use Storybook for component development
✅ **Do**: Review and approve all visual changes

## Integration with Design System

All components must:
- Use design tokens from the established design system
- Follow Tailwind CSS utility-first approach
- Implement proper motion system animations
- Support theme switching (light/dark modes)
- Maintain accessibility standards

## Enforcement

This workflow is enforced through:
1. **CI Pipeline**: Chromatic checks block PR merging
2. **Code Review**: Stories are reviewed alongside component code
3. **Visual Review**: All UI changes require visual approval
4. **Documentation**: This guide is part of required reading for all UI contributors

---

**Remember**: Storybook is not just documentation—it's the living specification of our UI components. Treating it as such ensures our interface remains consistent, reliable, and maintainable.