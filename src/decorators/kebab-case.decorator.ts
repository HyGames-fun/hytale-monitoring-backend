import { applyDecorators } from '@nestjs/common'
import { Validate } from 'class-validator'
import { IsKebabCaseValidator } from '../validators/kebab-case.validator'

export function IsKebabCase() {
  return applyDecorators(Validate(IsKebabCaseValidator))
}