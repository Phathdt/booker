#!/bin/bash
# Post-process generated schemas.ts:
# 1. Remove Partial<> TS type aliases (we use z.infer instead)
# 2. Remove explicit z.ZodType<> annotations (they force Partial on infer)
# 3. Add header comment

FILE="src/core/api/generated/schemas.ts"

# Remove all "type Dto/Market/V2... = Partial<{...}>;" blocks at the top
sed -i '' '/^type [A-Z].*= Partial<{$/,/^}>;$/d' "$FILE"

# Remove ": z.ZodType<...>" annotations from const declarations
sed -i '' 's/: z\.ZodType<[^>]*>//g' "$FILE"

# Add header
sed -i '' '1s/^/\/\/ AUTO-GENERATED — DO NOT EDIT (regenerate: pnpm generate:api)\n/' "$FILE"

echo "Post-processed $FILE"
