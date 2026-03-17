import { Module } from '@nestjs/common'
import { TurnstileService } from './turnstile.service'
import { TurnstileController } from './turnstile.controller'
import { HttpModule } from '@nestjs/axios'

@Module({
  imports: [HttpModule],
  controllers: [TurnstileController],
  providers: [TurnstileService],
  exports: [TurnstileService]
})
export class TurnstileModule {}
