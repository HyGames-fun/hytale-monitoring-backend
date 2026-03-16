import {
  CallHandler,
  ExecutionContext,
  Injectable,
  NestInterceptor
} from '@nestjs/common'
import { map, Observable } from 'rxjs'

@Injectable()
export class GlobalInterceptor implements NestInterceptor {
  intercept(context: ExecutionContext, next: CallHandler): Observable<any> {
    return next.handle().pipe(
      map((data) => {
        return {
          status: 'success',
          data: data as object
        }
      })
    )
  }
}
