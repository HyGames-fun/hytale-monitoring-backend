import {
  ArgumentsHost,
  Catch,
  ExceptionFilter,
  HttpException,
  HttpStatus,
  Logger
} from '@nestjs/common'
import { Response } from 'express'
import { ValidationException } from '../exceptions/validation.exception'

@Catch()
export class GlobalFilter<T> implements ExceptionFilter {
  private readonly logger = new Logger(GlobalFilter.name)

  catch(exception: T, ctx: ArgumentsHost) {
    const res: Response = ctx.switchToHttp().getResponse()
    //this.logger.error(exception, ctx)

    if (exception instanceof ValidationException) {
      res.status(exception.getStatus()).json(exception.getResponse())
      return
    }

    if (exception instanceof HttpException) {
      res.status(exception.getStatus()).json({
        status: 'error',
        message: exception.message,
        code: exception.getStatus()
      })
      return
    }

    res.status(HttpStatus.INTERNAL_SERVER_ERROR).json({
      status: 'error',
      message: 'Internal server error',
      code: HttpStatus.INTERNAL_SERVER_ERROR
    })
  }
}
