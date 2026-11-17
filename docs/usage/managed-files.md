# Editing GLK-managed YAML files

GLK updates managed files with a three-way merge between the previous generated
default (stored under `.glk/defaults/`), the newly generated default, and the
current file. User overrides are preserved while unchanged values receive
default updates. Keys deleted by the user stay deleted; keys removed from the
default are removed from the managed mapping.

## Key order

Existing mapping keys retain their current order, including user-added keys
placed between generated keys. New default keys are appended in their generated
order. This applies recursively, including to mappings matched inside lists.
On first generation, keys follow the generated default's order, which may be
alphabetical for content serialized from Go objects.

Reordering keys does not affect value merging. GLK still normalizes YAML
formatting, such as indentation, so preserving key order does not imply a
byte-for-byte preservation of the original file.

List items are matched by a unique `name`, `id`, or `key` field when available,
otherwise by position for equally sized lists of mappings, or by set comparison.
Modified lists retain current item order and append new default items; unchanged
lists follow the new default's item order.
