import { findAppIconConflicts } from '../src/appIcons'

const conflicts = findAppIconConflicts()
if (conflicts.length > 0) {
  console.error('AppIcon Conflicts Found:')
  conflicts.forEach((c) => console.error(c))
  process.exit(1)
} else {
  console.log('No AppIcon shadowing conflicts found.')
}
