import { Module } from '@nestjs/common'
import { ServerService } from './server.service'
import { ServerController } from './server.controller'
import { TurnstileModule } from '../turnstile/turnstile.module'

@Module({
  imports: [TurnstileModule],
  controllers: [ServerController],
  providers: [ServerService]
})
export class ServerModule {}
