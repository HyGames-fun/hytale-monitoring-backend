import { applyDecorators } from '@nestjs/common'
import { Validate } from 'class-validator'
import { IsIpValidator } from '../validators/ip.validator'

export function IsIp() {
  return applyDecorators(Validate(IsIpValidator))
}
