import { applyDecorators } from '@nestjs/common'
import { Validate } from 'class-validator'
import { IsTrimValidator } from '../validators/trim.validator'

export function IsTrim() {
  return applyDecorators(Validate(IsTrimValidator))
}
