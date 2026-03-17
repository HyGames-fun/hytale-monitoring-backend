import { createParamDecorator } from '@nestjs/common'
import { User } from '../../generated/prisma/client'
import { Request } from 'express'

export const Authorized = createParamDecorator((data: keyof User, context) => {
  const request: Request = context.switchToHttp().getRequest()
  const user = request.user

  return user ? (data ? user[data] : user) : undefined
})
